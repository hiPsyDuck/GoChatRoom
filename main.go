package main

import (
	"chatroom/ws"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	go ws.InitHub.Run()

	r := gin.Default()

	r.LoadHTMLGlob("view/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", "")
	})

	r.GET("/ws", func(c *gin.Context) {
		ws.Handle(c.Writer, c.Request)
	})

	r.Run(":8000")
}
