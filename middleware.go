package goin

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/zR-Zr/goin/interfaces"
	"github.com/zR-Zr/goin/pkg/zerrors"
)

type Middleware func(HandlerFunc) HandlerFunc

func Chain(outer Middleware, others ...Middleware) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		for i := len(others) - 1; i >= 0; i-- {
			next = others[i](next)
		}
		return outer(next)
	}
}

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
