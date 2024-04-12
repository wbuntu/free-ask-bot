package middleware

import (
	"net/http"
	"time"

	"gitbub.com/wbuntu/gin-template/internal/pkg/log"
	"gitbub.com/wbuntu/gin-template/internal/pkg/utils"
	"github.com/gin-gonic/gin"
)

const (
	// 请求ID
	GinCtxRequestID = "X-Request-Id"
	// 请求来源
	GinCtxFrom = "X-Request-From"
	// 租户ID
	GinCtxTenantID = "X-Tenant-ID"
)

func HeaderExtracter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求ID
		requestID := c.GetHeader(GinCtxRequestID)
		if len(requestID) == 0 {
			requestID = utils.UUID()
		}
		c.Set(GinCtxRequestID, requestID)
		// 请求来源
		from := c.GetHeader(GinCtxFrom)
		if len(from) == 0 {
			from = "unkonwn"
		}
		c.Set(GinCtxFrom, from)
		// 用户 ID
		tenantID := c.GetHeader(GinCtxTenantID)
		if len(tenantID) == 0 {
			tenantID = "unkonwn"
		}
		c.Set(GinCtxTenantID, tenantID)
	}
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()
		// set logger into request context
		ctx := log.S(c.Request.Context(), log.WithFields(log.Fields{
			"module":    "api",
			"tenantID":  c.GetString(GinCtxTenantID),
			"requestID": c.GetString(GinCtxRequestID),
			"from":      c.GetString(GinCtxFrom),
		}))
		c.Request = c.Request.WithContext(ctx)
		// excute next
		c.Next()
		// log request
		statusCode := c.Writer.Status()
		logger := log.G(c).WithFields(log.Fields{
			"latency":    time.Since(start), // time to process
			"clientIP":   c.ClientIP(),
			"method":     c.Request.Method,
			"dataLength": c.Writer.Size(),
		})
		if len(c.Errors) > 0 {
			logger.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
			return
		}
		if statusCode >= http.StatusInternalServerError {
			logger.Error(path)
		} else if statusCode >= http.StatusBadRequest {
			logger.Warn(path)
		} else {
			logger.Info(path)
		}
	}
}
