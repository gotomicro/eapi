package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()
	r.GET("/app-b/hello", handleHello)
}

func handleHello(c *gin.Context) {
	c.JSON(http.StatusOK, "world")
}
