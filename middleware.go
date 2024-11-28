package goin

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zR-Zr/goin/interfaces"
	"github.com/zR-Zr/goin/pkg/auth"
	"github.com/zR-Zr/goin/pkg/config"
	"github.com/zR-Zr/goin/pkg/zerrors"
	"golang.org/x/time/rate"
)

type MiddlewareFunc func(c *Context)

func adaptMiddleware(middleware MiddlewareFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("ctx").(*Context)
		middleware(ctx)
	}
}

func multiAdaptMiddleware(middleware ...MiddlewareFunc) []gin.HandlerFunc {
	var mds []gin.HandlerFunc
	for _, m := range middleware {
		mds = append(mds, adaptMiddleware(m))
	}
	return mds
}

// type Middleware func(HandlerFunc) HandlerFunc

// func Chain(outer Middleware, others ...Middleware) Middleware {
// 	return func(next HandlerFunc) HandlerFunc {
// 		for i := len(others) - 1; i >= 0; i-- {
// 			next = others[i](next)
// 		}
// 		return outer(next)
// 	}
// }

func ReplaceContextMiddleware(logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建自定义 Context 实例
		ctx := NewContext(c, logger)

		// 将自定义 Context 存储到 Gin Context 中
		c.Set("ctx", ctx)

		// 继续处理请求
		c.Next()
	}
}

func GlobalErrorHandlerMiddleware(logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := debug.Stack()

				// 记录错误
				logger.Error("panic recoverd", err, map[string]any{
					"stack":      string(stack),
					"request_id": c.MustGet("ctx").(*Context).RequestID(),
					"url":        c.Request.URL.String(),
					"method":     c.Request.Method,
					"ip":         c.ClientIP(),
				})

				// 返回错误响应
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    5000,
					"message": "内部错误",
					"data":    nil,
				})
			}
		}()

		c.Next()
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		if validationErr := zerrors.AsValidationError(err); validationErr != nil {
			c.JSON(validationErr.HTTPStatusCode(), gin.H{
				"code":    validationErr.Code.Code(),
				"message": validationErr.Code.Message(),
				"data":    validationErr.Fields,
			})
			return
		}

		if dbErr := zerrors.AsDatabaseError(err); dbErr != nil {
			c.JSON(dbErr.HTTPStatusCode(), gin.H{
				"code":    dbErr.Code.Code(),
				"message": dbErr.Code.Message(),
			})
			return
		}

		if customErr := zerrors.AsZError(err); customErr != nil {
			c.JSON(customErr.HTTPStatusCode(), gin.H{
				"code":    customErr.Code.Code(),
				"message": customErr.Code.Message(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    5000,
			"message": "内部错误",
		})

	}
}

// -------------------- Logging 中间件 --------------------

func LogginMiddleware() HandlerFunc {
	return func(c *Context) {
		// 前置逻辑
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		// 执行下一个中间件或处理器,请求处理
		c.Next()

		// 后置逻辑
		end := time.Now()
		latency := end.Sub(start)

		clientIP := c.ClientIP()
		statusCode := c.Writer.Status()
		statusColor := c.Context.GetString("status_color")

		c.Logger.Info(fmt.Sprintf("%s %3d %s %13v %15s %s %s",
			statusColor, statusCode, method, latency,
			clientIP, path, c.RequestID()),
			"latency", latency,
			"clicnet_ip", clientIP,
			"method", method,
			"path", path,
			"status_code", statusCode,
			"request_id", c.RequestID(),
		)
	}
}

// -------------------- CORS 中间件 --------------------
// CORSConfig CORS 配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	MaxAge           time.Duration
}

func NewCORS(cfg *CORSConfig) HandlerFunc {
	return func(c *Context) {
		// 设置 CORS 头信息
		c.Header("Access-Control-Allow-Origin", strings.Join(cfg.AllowOrigins, ","))
		c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ","))
		c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ","))
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 处理 OPTIONS 请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 调用下一个中间件或处理器
		c.Next()
	}
}

// -------------------- RateLimiter 中间件 --------------------

const (
	defaultReteLimit       = 10          // 默认每秒请求限制
	defaultReteBurst       = 100         // 令牌桶容量
	defaultCleanupInterval = time.Minute // 清理过期令牌的间隔
)

// func NewRateLimiter(opts ...RateLimiterOption) HandlerFunc {
// 	o := &rateLimiterOptions{
// 		limit:           defaultReteLimit,
// 		burst:           defaultReteBurst,
// 		cleanupInterval: defaultCleanupInterval,
// 	}

// 	for _, opt := range opts {
// 		opt(o)
// 	}

// 	// 创建限流器
// 	limiter := rate.NewLimiter(rate.Limit(o.limit), o.burst)

// 	// 清理过期令牌的定时器
// 	ticker := time.NewTicker(o.cleanupInterval)
// 	go func() {
// 		for range ticker.C {
// 		}
// 	}()

// 	return func(c *Context) {

// 	}
// }

// 定义限流器配置项
type RateLimiterOption func(*rateLimiterOptions)

type rateLimiterOptions struct {
	limit           rate.Limit    // 每秒请求数限制
	burst           int           // 令牌桶容量
	cleanupInterval time.Duration // 清理过期令牌的间隔
}

func WithLimit(limit rate.Limit) RateLimiterOption {
	return func(o *rateLimiterOptions) {
		o.limit = limit
	}
}

func WithBurst(burst int) RateLimiterOption {
	return func(o *rateLimiterOptions) {
		o.burst = burst
	}
}

func WithCleanupInterval(interval time.Duration) RateLimiterOption {
	return func(o *rateLimiterOptions) {
		o.cleanupInterval = interval
	}
}

// -------------------- 认证中间件 -------------------------

func AuthMiddleware(cfg *config.Config) HandlerFunc {
	return func(c *Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			return
		}

		// 解析 token
		jwtAuth := auth.NewJWTAuth(cfg)

		customClaims, err := jwtAuth.ParseToken(parts[1])
		if err != nil {
			return
		}

		// 将用户信息添加到 Context 中
		c.SetUser(&auth.JWTUser{
			ID:       customClaims.ID,
			Username: customClaims.Username,
			Type:     customClaims.Type,
		})

		c.Next() // 调用下一个中间件或处理器

	}
}
