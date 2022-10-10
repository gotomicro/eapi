package dto

type GoodCreateReq struct {
	Title    string   `json:"title"`    // 标签
	SubTitle string   `json:"subTitle"` // 名称
	Cover    string   `json:"cover"`
	Arr      []string `json:"arr"`
	ArrInt   []int    `json:"arrInt"`
}
