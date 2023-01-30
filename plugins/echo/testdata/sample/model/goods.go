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

// GenericTypeResponse used for testing generic type
type GenericTypeResponse[T any] struct {
	Data     T                     `json:"data"`
	Metadata map[string]any        `json:"metadata"`
	Value    *SampleGenericType[T] `json:"value"`
}

type SampleGenericType[T any] struct {
	Value                T                            `json:"value"`
	SelfRef              GenericTypeResponse[T]       `json:"selfRef"`
	List                 List[T]                      `json:"list"`
	Array                Array[T]                     `json:"array"`
	Map                  Map[int, T]                  `json:"map"`
	MultipleParamGeneric MultipleParamGeneric[int, T] `json:"multipleParamGeneric"`
}

type List[T any] []T

type Array[T any] [4]T

type Map[T comparable, V any] map[T]V

type MultipleParamGeneric[T, V any] struct {
	A T
	B V
}

type Error struct {
	Message string `json:"message"`
}
