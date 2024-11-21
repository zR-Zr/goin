package main

import "github.com/zR-Zr/goin"

func main() {
	server := goin.New()

	api := server.Group("/api")
	api.GET("/hello", func(c *goin.Context) {
		c.Success("hello", nil)
	})

	server.Run(":8081")
}
