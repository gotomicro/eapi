package shop

import (
	"net/http"

	"server/pkg/view"

	"github.com/gin-gonic/gin"
)

// GoodsCreate 创建商品
// @consume application/json
// @produce application/json
func GoodsCreate(c *gin.Context) {
	var req view.GoodsCreateReq
	if err := c.Bind(&req); err != nil {
		// 参数无效
		c.JSON(http.StatusBadRequest, view.ErrInvalidArgument)
		return
	}

	// Biz logic here ...

	var res view.GoodsCreateRes
	// 创建成功
	c.JSON(http.StatusOK, res)
}

// GoodsInfo 商品详情
// @consume application/json
// @produce application/json
func GoodsInfo(c *gin.Context) {
	guid := c.Param("guid")

	// get goods info by guid
	_ = guid

	c.JSON(http.StatusOK, view.GoodsInfoRes{})
}