package parser

import (
	"go/ast"

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
	ResFuncName string // 返回的函数名称 JSONOK, JSONListPage
	FuncComment string // 函数名称
	ReqParam    string // 用于模版处显示
	ResParam    string // 用于模板处显示
	ResIsArray  bool   // 返回的是否是数组
	ResType     int
	ReqSwagger  *spec.Schema
	ResSwagger  *spec.Schema
}

// C.JSONOK
// 1 无数据
// 2 返回的直接结构体
// 3 返回的直接指针结构体
// 4 返回的变量
// 5 返回的某个变量函数
// 5. xx := GetSomething()  xx.List
// 5. info := goodv1.GoodSku{}
// 5 c.JSONOK(info.SpecList)
type ResStruct struct {
	ReqResInfo
	//ResParam   string
	ReferParam string // 在4的时候，是引用的变量
	//ModuleName string // 在4的时候，是引用的变量，可能为别的包
	Type int
	// FuncName 名称 // JSONOK, JSONListPage
	FuncName  string
	FuncValue *ast.FuncDecl // 用于debug，哪个函数
}
