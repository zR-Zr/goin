package goin

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zR-Zr/goin/interfaces"
	"github.com/zR-Zr/goin/pkg/zlog"
)

type Server struct {
	engine *gin.Engine
	router *Router
	logger interfaces.Logger
}

func New() *Server {
	// 初始化日志
	var err error
	var logger interfaces.Logger

	logger, err = zlog.CreateLogger(
		zlog.WithLevel(zlog.DebugLevel),
		zlog.WithOutputInConsole(),
		zlog.WithSeparateErrorFile("error.log", 100, 30, 30),
		// zlog.WithFile("logs.log", 100, 30, 30),
	)
	if err != nil {
		panic(err)
	}

	engine := gin.Default()
	engine.Use(
		GlobalErrorHandlerMiddleware(logger),
		ReplaceContextMiddleware(logger),
	)

	router := NewRouter(engine)

	return &Server{
		engine: engine,
		router: router,
		logger: logger,
	}

}

func (s *Server) Run(addr string) {
	srv := &http.Server{
		Addr:    addr,
		Handler: s.engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Panic("listen: %s\n", err)
		}
	}()

	// 监听系统信号,优雅的关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 创建上下文,设置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server force shutdown:", err)
	}

	log.Println("Server exiting")
}

func (s *Server) Use(middlewares ...HandlerFunc) {
	s.engine.Use(MultiAdaptHandlerFunc(middlewares...)...)
}

func (s *Server) Group(relativePath string, handlers ...HandlerFunc) *Router {
	// 使用 Gin的 Group 方法创建路由分组
	group := s.engine.Group(relativePath, MultiAdaptHandlerFunc(handlers...)...)

	// 返回自定义 Router 实例, 以便继续添加路由
	return NewRouter(group)
}
