package controller

import "github.com/gin-gonic/gin"

type GoodsController struct{}

func NewGoodsController() *GoodsController {
	return &GoodsController{}
}

func (s *GoodsController) Delete(c *gin.Context) {
	// Goods Guid
	_ = c.Param("guid")
}
