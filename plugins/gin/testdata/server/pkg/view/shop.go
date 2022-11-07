package view

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

type GoodsInfoRes struct {
	Title    string `json:"title"`
	SubTitle string `json:"subTitle"`
	Cover    string `json:"cover"`
	Price    int64  `json:"price"`
}
