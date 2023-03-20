package router

import (
	"server/pkg/handler"
	"server/pkg/shop"

	"github.com/gin-gonic/gin"
)

type CustomGroup struct {
	*gin.RouterGroup
}

// ServeHttp
// @tags Shop
func ServeHttp() *gin.Engine {
	r := gin.Default()

	// custom router group
	cg := &CustomGroup{RouterGroup: &r.RouterGroup}
	cg.DELETE("/api/goods/:guid", handler.Handler(shop.GoodsDelete))

	// wrapped handler
	cg.GET("/wrapped-handler", handler.Handler(shop.WrappedHandler))

	g := r.Group("/api")

	// @tags Goods
	// @security oauth2 goods:read
	{
		// @security oauth2 goods:read goods:write
		{
			g.POST("/goods", handler.Handler(shop.GoodsCreate))
			g.POST("/goods/:guid/down", shop.GoodsDown)
		}
		g = g.Group("/v2")
		g.GET("/goods/:guid", shop.GoodsInfo)
	}

	return r
}
