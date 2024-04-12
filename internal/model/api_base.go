package model

import (
	"strings"

	"gitbub.com/wbuntu/gin-template/internal/pkg/log"
	"github.com/gin-gonic/gin"
)

type Route struct {
	Method     string
	Path       string
	Ctrl       Controller
	Middleware []gin.HandlerFunc
}

type Controller interface {
	// 公共接口
	Init()                // 构造函数
	SetLogger(log.Logger) // 设置logger
	SkipIO() bool         // 关闭自动反序列化和写回响应
	// 控制器接口
	Input() Request     // 请求
	Output() Response   // 响应
	Serve(*gin.Context) // 控制器逻辑
}

type BaseController struct {
	Logger log.Logger
}

func (c *BaseController) Init() {}

func (c *BaseController) SetLogger(logger log.Logger) {
	c.Logger = logger
}

func (c *BaseController) SkipIO() bool {
	return false
}

type Request interface {
	SetTenantID(string)
	GetTenantID() string
	SetRequestID(string)
	GetRequestID() string
	SetFrom(string)
	GetFrom() string
}

type Response interface {
	Init()
	GetCode() Code
	GetMessage() string
	GetRequestID() string
	SetRequestID(string)
}

type BaseRequest struct {
	TenantID  string `json:"-"` // 租户ID
	RequestID string `json:"-"` // 请求ID
	From      string `json:"-"` // 请求来源
}

func (r *BaseRequest) SetTenantID(item string) {
	r.TenantID = item
}

func (r *BaseRequest) GetTenantID() string {
	return r.TenantID
}

func (r *BaseRequest) SetRequestID(item string) {
	r.RequestID = item
}

func (r *BaseRequest) GetRequestID() string {
	return r.RequestID
}

func (r *BaseRequest) SetFrom(item string) {
	r.From = item
}

func (r *BaseRequest) GetFrom() string {
	return r.From
}

type BaseResponse struct {
	Code      Code   `json:"code" example:"Success"`                                             // 响应码
	Message   string `json:"message" example:"调用成功"`                                             // 响应消息
	RequestID string `json:"requestId,omitempty" example:"6893b1e9-da8f-4c6c-a161-eba4b81ea5b3"` // 请求ID
}

func (r *BaseResponse) Init() {
	r.Code = RespSuccess.Code
	r.Message = RespSuccess.Message
}

func (r *BaseResponse) GetCode() Code {
	return r.Code
}

func (r *BaseResponse) GetMessage() string {
	return r.Message
}

func (r *BaseResponse) GetRequestID() string {
	return r.RequestID
}

func (r *BaseResponse) SetRequestID(item string) {
	r.RequestID = item
}

// Update 更新response的code并设置默认message，若传入的messages不为空，使用:分隔拼接附加到response的message中
func (r *BaseResponse) Update(code Code, messages ...string) {
	r.Code = code
	r.Message = code.Message()
	if len(messages) > 0 {
		r.Message += ": " + strings.Join(messages, ": ")
	}
}
