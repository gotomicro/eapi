package parser

import (
	"go/ast"
	"net/http"

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

func (i *UrlInfo) GetOperationSpec() *spec.Operation {
	operation := spec.NewOperation(i.UniqueKey)
	operation.WithDescription(i.FuncComment)

	// TODO 支持 Query/FormData 参数

	if i.ReqSwagger != nil {
		param := &spec.Parameter{}
		param.Schema = i.ReqSwagger
		param.Name = i.ReqParam
		// FIXME c.Bind() 方法无法直接知道来源的参数类型是 query/form-data/json-in-body/...
		// c.Bind() 内部的逻辑是根据 Content-Type 和 HTTP Method 共同决定从哪个参数里面解析. 注释原文:
		//		"Bind checks the Method and Content-Type to select a binding engine automatically"
		// 然而， Content-Type 在解析代码的时候是未知的，因此我们无法还原编程人员的真实意图
		param.In = "body"
		operation.AddParam(param)
	}
	if i.ResSwagger != nil {
		operation.RespondsWith(http.StatusOK, spec.NewResponse().WithSchema(i.ResSwagger))
	}

	return operation
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
