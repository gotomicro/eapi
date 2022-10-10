package router

import (
	"bff/pkg/bffcore"
	"bff/pkg/shop"
	"github.com/gotomicro/ego/server/egin"
)

func ServeHttp() *egin.Component {
	r := egin.Load("server.http").Build()
	g := r.Group("/api")
	g = g.Group("/test")
	g.POST("/new/goodCreate", bffcore.Handle(shop.GoodCreate)) // 微信通知回调接口
	//g.POST("/new/goodCreate2", bffcore.Handle(shop.ListArticle)) // 微信通知回调接口
	//g.POST("/new/gooddto", shop.GoodCreateTest)                // 微信通知回调接口
	//g.POST("/test", shop.Test)                                 // 微信通知回调接口
	return r
}
