package goin

import "github.com/gin-gonic/gin"

type HandlerFunc func(c *Context)

func AdaptHandlerFunc(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("ctx").(*Context)
		handler(ctx)
	}
}

func MultiAdaptHandlerFunc(handlers ...HandlerFunc) []gin.HandlerFunc {
	var ginHandlers []gin.HandlerFunc

	for _, handler := range handlers {
		ginHandlers = append(ginHandlers, AdaptHandlerFunc(handler))
	}

	return ginHandlers
}
