package api

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"gitbub.com/wbuntu/free-ask-bot/internal/api/middleware"
	"gitbub.com/wbuntu/free-ask-bot/internal/model"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/config"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/log"
)

func (s *Server) setupEngine(ctx context.Context, c *config.Config) error {
	// 设置模式
	gin.SetMode(gin.ReleaseMode)
	// 初始化
	g := gin.New()
	// 启动Context特性
	g.ContextWithFallback = true
	// 配置中间件
	g.Use(
		gin.Recovery(),
	)
	// healthz && pprof && swagger && metrics
	addCommonRoutes(g)
	// controller
	addControllerRoute(g)
	// srv
	s.srv = &http.Server{
		Addr:    c.API.Addr,
		Handler: g,
	}
	// tlsSrv
	if len(c.API.TLSAddr) > 0 && len(c.API.TLSCrt) > 0 && len(c.API.TLSCrt) > 0 {
		s.tlsSrv = &http.Server{
			Addr:    c.API.TLSAddr,
			Handler: g,
		}
	}
	return nil
}

func addCommonRoutes(g *gin.Engine) {
	// readyz
	g.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, model.RespSuccess)
	})
	// healthz
	g.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, model.RespSuccess)
	})
	// pprof
	pprof.Register(g)
	// swagger
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, func(c *ginSwagger.Config) {
		c.DefaultModelsExpandDepth = 10
	}))
}

func addControllerRoute(g *gin.Engine) {
	v1 := g.Group("/api/v1.0")
	// 配置中间件
	v1.Use(
		middleware.HeaderExtracter(),
		middleware.RequestLogger(),
		gzip.Gzip(gzip.DefaultCompression),
	)
	// 配置 routes
	funcMap := map[string]func(string, ...gin.HandlerFunc) gin.IRoutes{
		http.MethodGet:     v1.GET,
		http.MethodPost:    v1.POST,
		http.MethodDelete:  v1.DELETE,
		http.MethodPatch:   v1.PATCH,
		http.MethodPut:     v1.PUT,
		http.MethodOptions: v1.OPTIONS,
		http.MethodHead:    v1.HEAD,
		"ANY":              v1.Any,
	}
	routes := getRoutes()
	for i := range routes {
		v := routes[i]
		fn, ok := funcMap[v.Method]
		if !ok {
			continue
		}
		v.Middleware = append(v.Middleware, handlerFuncWrapper(v.Ctrl))
		fn(v.Path, v.Middleware...)
	}
}

func getRoutes() []model.Route {
	routes := []model.Route{}
	routes = append(routes, clusterRoute...)
	routes = append(routes, utilsRoute...)
	return routes
}

func handlerFuncWrapper(ctrlTpl model.Controller) gin.HandlerFunc {
	// 获取controller类型
	ctrlType := reflect.TypeOf(ctrlTpl).Elem()
	return func(c *gin.Context) {
		// 根据类型创建实例
		ctrlInstance := reflect.New(ctrlType).Interface().(model.Controller)
		// 构造函数
		ctrlInstance.Init()
		// 配置logger
		logger := log.G(c)
		ctrlInstance.SetLogger(logger)
		// 初始化input
		input := ctrlInstance.Input()
		// 配置公共字段
		input.SetTenantID(c.GetString(middleware.GinCtxTenantID))
		input.SetRequestID(c.GetString(middleware.GinCtxRequestID))
		input.SetFrom(c.GetString(middleware.GinCtxFrom))
		// 初始化output
		ctrlInstance.Output().Init()
		// 设置requestID
		ctrlInstance.Output().SetRequestID(input.GetRequestID())
		// 跳过控制器逻辑、数据序列化和反序列化，由controller直接处理，例如Websocket、代理等
		if ctrlInstance.SkipIO() {
			ctrlInstance.Serve(c)
			return
		}
		// 根据请求方法选择binding
		var bindFn func(interface{}) error
		// POST、PATCH、PUT：从Body中解析JSON
		// 其他：从Query中解析FORM，参数值禁止嵌套，不允许使用数组、字典和结构体
		switch c.Request.Method {
		case http.MethodPost, http.MethodPatch, http.MethodPut:
			bindFn = c.ShouldBindJSON
		case http.MethodGet, http.MethodDelete, http.MethodOptions, http.MethodHead:
			bindFn = c.ShouldBindQuery
		}
		// 解析Body和Query 反序列化 + 参数检查
		if err := bindFn(input); err != nil {
			logger.Errorf("unmarshal and validate input: %s", err)
			c.JSON(http.StatusOK, model.BaseResponse{
				Code:    model.CodeParamError,
				Message: fmt.Sprintf("param error: %s", err),
			})
			return
		}
		// 执行控制器逻辑
		ctrlInstance.Serve(c)
		// 获取output结构体指针
		output := ctrlInstance.Output()
		// JSON格式响应
		c.JSON(http.StatusOK, output)
	}
}
