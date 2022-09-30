package router

import (
	"bff/pkg/shop"
	"github.com/gotomicro/ego/server/egin"
)

func ServeHttp() *egin.Component {
	r := egin.Load("server.http").Build()
	g := r.Group("/api")
	g = g.Group("/test")
	g.POST("/new/goodCreate", shop.GoodCreate) // 微信通知回调接口
	g.POST("/test", shop.Test)                 // 微信通知回调接口
	return r
}
