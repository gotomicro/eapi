package shop

import (
	"github.com/gin-gonic/gin"
)

func Test(c *gin.Context) {
	var req GoodCreateReq
	if err := c.Bind(&req); err != nil {
		c.JSON(1, struct{}{})
		return
	}
	c.JSON(200, req)
}
