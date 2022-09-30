package shop

import (
	"github.com/gin-gonic/gin"
)

type GoodCreateReq struct {
	Title    string   `json:"title"`    // 标签
	SubTitle string   `json:"subTitle"` // 名称
	Cover    string   `json:"cover"`
	Arr      []string `json:"arr"`
}

// GoodCreate 创建商品
func GoodCreate(c *gin.Context) {
	var req GoodCreateReq
	if err := c.Bind(&req); err != nil {
		c.JSON(1, struct{}{})
		return
	}
	c.JSON(200, req)
}
