package view

import (
	"encoding/json"

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

type SelfRefType struct {
	Data   string       `json:"data"`
	Parent *SelfRefType `json:"parent"`
}

type StringAlias = string

type GoodsCreateRes struct {
	// 商品 GUID
	Guid string `json:"guid"`
	// 测试循环引用
	SelfRef *SelfRefType `json:"selfRef"`
	// 测试引用第三方包
	Status gin.Params
	// 测试类型别名
	StringAlias StringAlias `json:"stringAlias"`
	// 测试引用内置包类型
	Raw json.RawMessage `json:"raw"`
}

type GoodsInfoRes struct {
	Title    string `json:"title"`
	SubTitle string `json:"subTitle"`
	Cover    string `json:"cover"`
	Price    int64  `json:"price"`
}

type GoodsDownRes struct {
	Status string `json:"status"`
}
