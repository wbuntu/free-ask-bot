package tools

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"gitbub.com/wbuntu/gin-template/internal/model"
	"github.com/pkg/errors"
)

type HTTPContextKey string

const (
	// CustomResolveKey 提供类似配置 hosts 的功能，实现简单的自定义域名解析
	CustomResolveKey HTTPContextKey = "CustomResolveKey"
	// CustomHeaderKey 提供自定义请求头的功能，与每个 HTTP 请求关联
	CustomHeaderKey HTTPContextKey = "CustomHeaderKey"
)

// NewHTTPClient 根据提供的参数构建专用的HTTPClient
func NewHTTPClient(baseURL string, header map[string]string) *HTTPClient {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 15 * time.Second,
	}
	return &HTTPClient{
		client: &http.Client{
			Transport: &http.Transport{
				// 忽略证书验证，优先使用服务器偏好的加密套件
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify:       true,
					PreferServerCipherSuites: true,
				},
				// 配置Dialer
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					// 在初始化连接阶段检查是否需要使用自定义hosts
					if customResolveIP := ctx.Value(CustomResolveKey); customResolveIP != nil {
						if v, ok := customResolveIP.(*string); ok {
							_, port, _ := net.SplitHostPort(addr)
							addr = net.JoinHostPort(*v, port)
						}
					}
					return dialer.DialContext(ctx, network, addr)
				},
				// 默认启用HTTP2
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   100,
				MaxConnsPerHost:       100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			// 配置请求超时
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
		header:  header,
		// 重试次数默认为3
		retryCount: 3,
		// 重试间隔默认为200毫秒
		retryWait: time.Millisecond * 200,
	}
}

// HTTPClient 专用于调用接口的HTTP客户端
type HTTPClient struct {
	client     *http.Client
	baseURL    string
	header     map[string]string
	retryCount int
	retryWait  time.Duration
}

func (c *HTTPClient) POST(ctx context.Context, path string, request model.Request, response model.Response) error {
	return c.do(ctx, http.MethodPost, path, request, response)
}

func (c *HTTPClient) PUT(ctx context.Context, path string, request model.Request, response model.Response) error {
	return c.do(ctx, http.MethodPut, path, request, response)
}

func (c *HTTPClient) GET(ctx context.Context, path string, request model.Request, response model.Response) error {
	return c.do(ctx, http.MethodGet, path, request, response)
}

func (c *HTTPClient) DELETE(ctx context.Context, path string, request model.Request, response model.Response) error {
	return c.do(ctx, http.MethodDelete, path, request, response)
}

func (c *HTTPClient) do(ctx context.Context, method string, path string, request model.Request, response model.Response) error {
	var (
		req  *http.Request
		resp *http.Response
		err  error
	)
	// 序列化应用层请求
	data, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "marshal request")
	}
	// 构造 HTTP 请求
	url := c.baseURL + path
	switch method {
	case http.MethodPost, http.MethodPatch, http.MethodPut:
		// 参数放入Body
		req, err = http.NewRequestWithContext(
			ctx,
			method,
			url,
			bytes.NewBuffer(data),
		)
	case http.MethodGet, http.MethodDelete, http.MethodOptions, http.MethodHead:
		// 参数放入URL 注意：嵌套参数不做展开
		m := map[string]interface{}{}
		if err := json.Unmarshal(data, &m); err != nil {
			return errors.Wrapf(err, "convert request")
		}
		if len(m) > 0 {
			params := []string{}
			for k, v := range m {
				params = append(params, fmt.Sprintf("%s=%v", k, v))
			}
			url += "?" + strings.Join(params, "&")
		}
		req, err = http.NewRequestWithContext(
			ctx,
			method,
			url,
			nil,
		)
	default:
		return errors.Errorf("unsupported method: %s", method)
	}
	if err != nil {
		return errors.Wrap(err, "build http request")
	}
	// 设置 Content-Type 与 Accept 头部
	switch req.Method {
	case http.MethodPost, http.MethodPut:
		// 只有POST和PUT请求可以设置Content-Type头部
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	// 设置默认自定义头部
	if len(c.header) > 0 {
		for k, v := range c.header {
			req.Header.Set(k, v)
		}
	}
	// 获取 HTTP 响应，出错时自动重试
	for i := 0; i < c.retryCount; i++ {
		resp, err = c.client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(c.retryWait)
	}
	if err != nil {
		return errors.Wrap(err, "send http request")
	}
	// 读取 body
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "read http response")
	}
	// 检查 HTTP 层响应码
	if resp.StatusCode != http.StatusOK {
		errStr := fmt.Sprintf("%s: %s: status: %s:", method, req.URL, resp.Status)
		if len(respData) > 0 {
			errStr += fmt.Sprintf(" msg: %s", string(respData))
		}
		return errors.New(errStr)
	}
	// 反序列化应用层响应
	if err := json.Unmarshal(respData, response); err != nil {
		return errors.Wrapf(err, "unmarshal response: %s", string(respData))
	}
	// 检查应用层返回码
	if response.GetCode() != model.CodeSuccess {
		return errors.Errorf("response: code: %s message: %s", response.GetCode(), response.GetMessage())
	}
	return nil
}
