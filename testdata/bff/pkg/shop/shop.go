package shop

import (
	"bff/pkg/bffcore"
)

type GoodCreateReq struct {
	Title    string   `json:"title"`    // 标签
	SubTitle string   `json:"subTitle"` // 名称
	Cover    string   `json:"cover"`
	Arr      []string `json:"arr"`
	ArrInt   []int    `json:"arrInt"`
	//Teacher  Teacher           `json:"teacher"`
}

//
//type Teacher struct {
//	Title string `json:"title"` // 标签
//
//}
//
type TeacherRes struct {
	Title string `json:"title"` // 标签
}

func GoodCreate(c *bffcore.Context) {
	//var req GoodCreateReq
	//if err := c.Bind(&req); err != nil {
	//	c.JSONE(1, "error", nil)
	//	return
	//}
	////info, _ := invoker.Grpc.ListAllByUid(c.Request.Context(), &communityv1.ListAllByUidRequest{})
	res := []TeacherRes{}
	c.JSONOK(res, nil)

}

// GoodCreate 创建商品
//func GoodCreateTest(c *gin.Context) {
//	var req dto.GoodCreateReq
//	if err := c.Bind(&req); err != nil {
//		c.JSON(1, struct{}{})
//		return
//	}
//	c.JSON(200, req)
//}

//
//// GoodCreate 创建商品
//func GoodCreateTest(c *gin.Context) {
//	var req ReferralsSendReq
//	if err := c.Bind(&req); err != nil {
//		c.JSON(1, struct{}{})
//		return
//	}
//	c.JSON(200, req)
//}
