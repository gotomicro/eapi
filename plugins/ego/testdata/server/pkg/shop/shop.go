package shop

import (
	"net/http"

	"server/pkg/view"

	"github.com/gin-gonic/gin"
)

type Image struct {
	// 图片链接
	// @required
	Url string `json:"url"`
}

type GoodsCreateReq struct {
	// 商品标题
	// @required
	Title string `json:"title"`

	// 商品描述
	// @nullable
	SubTitle string `json:"subTitle"`

	// 封面图
	Cover string `json:"cover"`

	// 价格(分)
	// @required
	Price int64 `json:"price"`

	// 详情图
	// @nullable
	Images []*Image `json:"images"`
}

type GoodsCreateRes struct {
	// 商品 GUID
	Guid string `json:"guid"`
}

// GoodsCreate 创建商品
// @consume application/json
// @produce application/json
func GoodsCreate(c *gin.Context) {
	var req GoodsCreateReq
	if err := c.Bind(&req); err != nil {
		// 参数无效
		c.JSON(http.StatusBadRequest, view.ErrInvalidArgument)
		return
	}

	// Biz logic here ...

	var res GoodsCreateRes
	// 创建成功
	c.JSON(http.StatusOK, res)
}

type GoodsInfoRes struct {
	Title    string `json:"title"`
	SubTitle string `json:"subTitle"`
	Cover    string `json:"cover"`
	Price    int64  `json:"price"`
}

func GoodsInfo(c *gin.Context) {
	guid := c.Param("guid")

	// get goods info by guid
	_ = guid

	c.JSON(http.StatusOK, GoodsInfoRes{})
}
