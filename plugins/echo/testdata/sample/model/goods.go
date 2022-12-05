package model

type GoodsStatus int

const (
	GoodsOnSale GoodsStatus = iota + 1
	GoodsOffSale
	GoodsOutOfStock
)

type GoodsInfo struct {
	// Unique key
	ID     int64       `json:"id"`
	Title  string      `json:"title"`
	Status GoodsStatus `json:"status"`
	Stock  int         `json:"stock"`
	// Url of cover image
	Cover string `json:"cover"`
	// Detail images
	Images []*Image `json:"images"`
}

type Image struct {
	Src   string `json:"src"`
	Title string `json:"title"`
}

type ListGoodsRequest struct {
	// Obtained in `nextPage` field of previous page response.
	Since string `query:"since"`
	// @required
	Limit int64 `query:"limit"`
}

type ListGoodsResponse struct {
	Items []*GoodsInfo `json:"items"`
	// Url of next page. If there is no more items, nextPage field not exists.
	NextPage string `json:"nextPage,omitempty"`
}

type CreateGoodsRequest struct {
	Title  string      `json:"title"`
	Status GoodsStatus `json:"status"`
	Stock  int         `json:"stock"`
	// Url of cover image
	Cover string `json:"cover"`
	// Detail images
	Images []*Image `json:"images"`
}

type UpdateGoodsRequest struct {
	Title  *string      `json:"title"`
	Status *GoodsStatus `json:"status"`
	Stock  *int         `json:"stock"`
	// Url of cover image
	Cover *string `json:"cover"`
	// Detail images
	Images *[]*Image `json:"images"`
}
