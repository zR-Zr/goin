package goin

import "github.com/gin-gonic/gin"

// Router 自定义 Router
type Router struct {
	router gin.IRouter
}

func NewRouter(router gin.IRouter) *Router {
	return &Router{router: router}
}

func (r *Router) POST(path string, handlers ...HandlerFunc) {
	r.router.POST(path, MultiAdaptHandlerFunc(handlers...)...)
}

func (r *Router) GET(path string, handlers ...HandlerFunc) {
	r.router.GET(path, MultiAdaptHandlerFunc(handlers...)...)
}

func (r *Router) PUT(path string, handlers ...HandlerFunc) {
	r.router.PUT(path, MultiAdaptHandlerFunc(handlers...)...)
}

func (r *Router) DELETE(path string, handlers ...HandlerFunc) {
	r.router.DELETE(path, MultiAdaptHandlerFunc(handlers...)...)
}

func (r *Router) Use(middleware ...HandlerFunc) {
	ginHandlers := MultiAdaptHandlerFunc(middleware...)
	r.router.Use(ginHandlers...)
}

func wrapper(hs ...HandlerFunc) []gin.HandlerFunc {
	var ginHandlers []gin.HandlerFunc

	for _, f := range hs {
		ginHandlers = append(ginHandlers, AdaptHandlerFunc(f))
	}

	return ginHandlers
}
