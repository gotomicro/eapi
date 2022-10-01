package parser

import (
	"github.com/go-openapi/spec"
)

type UrlInfo struct {
	CurrentMod  string
	UniqueKey   string
	FullPath    string
	Method      string
	Prefix      string
	PackagePath string
	ModuleName  string
	FuncName    string
	ReqParam    string
	ResParam    string
	ResType     int
	ReqSwagger  *spec.Schema
	ResSwagger  *spec.Schema
}

// C.JSONOK
// 1 无数据
// 2 返回的直接结构体
// 3 返回的直接指针结构体
// 4 返回的变量
type ResStruct struct {
	ReqResInfo
	//ResParam   string
	ReferParam string // 在4的时候，是引用的变量
	//ModuleName string // 在4的时候，是引用的变量，可能为别的包
	Type int
}
