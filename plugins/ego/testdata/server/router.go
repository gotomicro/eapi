package router

import (
	"server/pkg/shop"

	"github.com/gotomicro/ego/server/egin"
)

func ServeHttp() *egin.Component {
	r := egin.Load("server.http").Build()
	g := r.Group("/api")
	g.POST("/goods", shop.GoodsCreate)
	g = g.Group("/v2")
	g.GET("/goods/:guid", shop.GoodsInfo)
	return r
}
