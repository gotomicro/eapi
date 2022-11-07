package router

import (
	"server/pkg/shop"

	"github.com/gin-gonic/gin"
)

func ServeHttp() *gin.Engine {
	r := gin.Default()
	g := r.Group("/api")
	g.POST("/goods", shop.GoodsCreate)
	g = g.Group("/v2")
	g.GET("/goods/:guid", shop.GoodsInfo)
	return r
}
