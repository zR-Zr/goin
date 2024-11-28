package main

import "github.com/zR-Zr/goin"

func main() {
	server := goin.NewServer("config.yaml")
	server.Use(goin.LogginMiddleware())

	api := server.Group("/api")
	api.GET("/hello", func(c *goin.Context) {
		c.Success("hello", nil)
	})

	server.Run(":8081")
}
