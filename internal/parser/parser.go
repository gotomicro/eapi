package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/gotomicro/ego/core/elog"
	"github.com/samber/lo"
)

var ErrRecursiveParseStruct = errors.New("recursively parsing struct")

type Swagger struct {
	spec.SwaggerProps
}

type astParser struct {
	readContent string // 读取原文件数据
	userOption  UserOption
	GoMod       string
	// ParseDependencies whether swag should be parse outside dependency folder
	ParseAllDependency bool
	Dependences        []string
	ResFuncs           []string
	importMap          map[string]ImportInfo
	routerFunc         *ast.FuncDecl
	routerParam        string // router的变量名
	// 存储group信息
	groupMap map[string]GroupInfo
	// 存储url信息
	urlMap map[string]UrlInfo
	// 顺序URL
	sortedUrl []UrlInfo
	// packages store entities of APIs, definitions, file, package path etc.  and their relations
	packages *PackagesDefinitions
	// ParseInternal whether ego-gen-api should parse internal packages
	ParseInternal bool
	// excludes excludes dirs and files in SearchDir
	excludes map[string]struct{}
	// ParseVendor parse vendor folder
	ParseVendor bool

	// structStack stores full names of the structures that were already parsed or are being parsed now
	structStack []*TypeSpecDef
	// fieldParserFactory create FieldParser
	fieldParserFactory FieldParserFactory
	// PropNamingStrategy naming strategy
	PropNamingStrategy string

	// parsedSchemas store schemas which have been parsed from ast.TypeSpec
	parsedSchemas map[*TypeSpecDef]*Schema

	// outputSchemas store schemas which will be export to swagger
	outputSchemas map[*TypeSpecDef]*Schema

	// existSchemaNames store names of models for conflict determination
	existSchemaNames map[string]*Schema

	// toBeRenamedSchemas names of models to be renamed
	toBeRenamedSchemas map[string]string

	// toBeRenamedSchemas URLs of ref models to be renamed
	toBeRenamedRefURLs []*url.URL

	swagger *spec.SwaggerProps
}

// FieldParserFactory create FieldParser
type FieldParserFactory func(ps *astParser, field *ast.Field) FieldParser

// FieldParser parse struct field
type FieldParser interface {
	ShouldSkip() (bool, error)
	FieldName() (string, error)
	CustomSchema() (*spec.Schema, error)
	ComplementSchema(schema *spec.Schema) error
	IsRequired() (bool, error)
}

type ImportInfo struct {
	KeyName     string
	Alias       string
	PackagePath string
}

type GroupInfo struct {
	ParamName string // 变量名
	Prefix    string // 前缀
}

func AstParserBuild(userOption UserOption) (*astParser, error) {
	a := &astParser{
		readContent:        "",
		userOption:         userOption,
		importMap:          map[string]ImportInfo{},
		routerParam:        "",
		groupMap:           map[string]GroupInfo{},
		urlMap:             map[string]UrlInfo{},
		sortedUrl:          make([]UrlInfo, 0),
		fieldParserFactory: newTagBaseFieldParser,
		packages:           NewPackagesDefinitions(),
		swagger: &spec.SwaggerProps{
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Contact: &spec.ContactInfo{},
					License: nil,
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{},
				},
			},
			Paths: &spec.Paths{
				Paths: make(map[string]spec.PathItem),
			},
			Definitions:         make(map[string]spec.Schema),
			SecurityDefinitions: make(map[string]*spec.SecurityScheme),
		},
		parsedSchemas:      make(map[*TypeSpecDef]*Schema),
		outputSchemas:      make(map[*TypeSpecDef]*Schema),
		existSchemaNames:   make(map[string]*Schema),
		toBeRenamedSchemas: make(map[string]string),
	}

	// 初始化配置，校验数据
	a.initOption()

	err := a.ParseAPI(a.userOption.RootPath, a.userOption.RootMainGo, 100)
	if err != nil {
		panic("parse api fail" + err.Error())
	}

	// 解析group，url
	a.parserStruct()

	// {FullPath:/api/test/new/goodCreate Method:POST Prefix:/api/test ModuleName:shop FuncName:GoodCreate}
	// package path: File:0x140003fa200 Path:/Users/askuy/code/github/gotomicro/ego-gen-api/internal/parser/testdata/bff/pkg/shop/shop.go ModName:bff/pkg/shop
	for _, urlInfo := range a.sortedUrl {
		a.packages.rangeByPkgPath(urlInfo.PackagePath, func(filename string, currentFile *ast.File) error {
			var handlerFnDecl *ast.FuncDecl

			// 先找到有没有名称为value.FuncName的函数
			for _, declValue := range currentFile.Decls {
				// 找到函数
				fnDecl, flag := declValue.(*ast.FuncDecl)
				if !flag {
					continue
				}
				if fnDecl.Name.String() != urlInfo.FuncName {
					continue
				}

				handlerFnDecl = fnDecl
				break
			}

			if handlerFnDecl != nil {
				parseHandlerFn(handlerFnDecl, urlInfo, a, currentFile)
			}
			return nil
		})
	}

	return a, nil
}

// 解析 handler 函数
func parseHandlerFn(handlerFnDecl *ast.FuncDecl, urlInfo UrlInfo, a *astParser, currentFile *ast.File) {
	var comment string
	var bindReqName string

	if handlerFnDecl.Doc != nil {
		for _, commentInfo := range handlerFnDecl.Doc.List {
			comment += commentInfo.Text + "\n"
		}
	}
	comment = strings.TrimSuffix(comment, "\n")
	urlInfo.FuncComment = comment

	// 获取参数名称，拿到他是(c *gin.Context)，还是(ctx *gin.Context)
	// 目前只存在一个参数
	if len(handlerFnDecl.Type.Params.List) != 1 {
		panic("参数个数不为1")
	}

	// 获取 context 参数名称
	ctxName := handlerFnDecl.Type.Params.List[0].Names[0].String()

	// 解析请求参数
	{
		bindReqName = parseRequestParamBinding(handlerFnDecl, ctxName)
		// warning 有这个参数，才需要去找下这个变量
		if bindReqName != "" {
			// 根据 req或者res的变量名找到对应的schema信息
			reqResInfo := a.getStructByReferName(ParamInfo{
				SelectorName: bindReqName,
			}, currentFile, handlerFnDecl.Body.List)
			if reqResInfo.StructName == "" {
				panic("reqResInfo.SelectorName is empty, " + bindReqName)
			}

			schema, reqResFullName := a.getReqResSchema(reqResInfo, urlInfo)
			urlInfo.ReqSwagger = schema
			urlInfo.ReqParam = reqResFullName
			a.urlMap[urlInfo.UniqueKey] = urlInfo
		}
	}

	resList := make([]ResStruct, 0)

	// 遍历取最后一次CallExpr，说明是 c.JSONOK
	ast.Inspect(handlerFnDecl, func(n ast.Node) bool {
		if !handlerFnDecl.Name.IsExported() {
			return true
		}
		switch nn := n.(type) {
		case *ast.CallExpr:
			fun, flag := nn.Fun.(*ast.SelectorExpr)
			if !flag {
				return true
			}
			funcXIndent, flag := fun.X.(*ast.Ident)
			if !flag {
				return true
			}

			//spew.Dump(nn)
			// 说明找到了c.JSONOK
			//c.JSONOK(GoodCreateReq{})
			//c.JSONOK(req)
			//c.JSONOK()
			if !(lo.Contains(a.ResFuncs, fun.Sel.Name) && funcXIndent.String() == ctxName) {
				return true
			}

			responseList := parseHandlerResponse(fun, handlerFnDecl, nn, currentFile)
			resList = append(resList, responseList...)
		}
		return true
	})

	// 取最后一次的数据
	if len(resList) > 0 {
		resInfo := resList[len(resList)-1]
		parseResponseCall(handlerFnDecl, urlInfo, a, currentFile, resInfo, bindReqName)
	}
}

// 解析 handler 内的输出响应的函数调用
func parseResponseCall(handlerFnDecl *ast.FuncDecl, urlInfo UrlInfo, a *astParser, currentFile *ast.File, resInfo ResStruct, bindReqName string) {
	urlInfo.ResType = resInfo.Type
	urlInfo.ResFuncName = resInfo.FuncName
	var reqResInfo ReqResInfo
	//fmt.Printf("urlInfo.ResFuncName--------------->"+"%+v\n", urlInfo.ResFuncName)
	var schema *spec.Schema
	var reqResFullName string
	switch resInfo.Type {
	case 2:
		schema, reqResFullName = a.getReqResSchema(resInfo.ReqResInfo, urlInfo)
	case 3:
		schema, reqResFullName = a.getReqResSchema(resInfo.ReqResInfo, urlInfo)
	case 4:
		// 根据 req或者res的变量名找到对应的struct name信息
		reqResInfo = a.getStructByReferName(ParamInfo{
			Type:         1,
			SelectorName: resInfo.ReferParam,
		}, currentFile, handlerFnDecl.Body.List)
		if reqResInfo.StructName == "" {
			fmt.Printf("resInfo--------------->"+"%+v\n", resInfo)
			fmt.Printf("resInfo--------------->"+"%+v\n", reqResInfo)
			//spew.Dump(resInfo)
			//panic(111)
			break
		}

		schema, reqResFullName = a.getReqResSchema(reqResInfo, urlInfo)
	case 5:
		fmt.Printf("resInfo--------------->"+"%+v%+v\n", resInfo)
		fmt.Printf("bindReqName--------------->"+"%+v\n", bindReqName)
		reqResInfo = a.getStructByReferName(ParamInfo{
			Type:         1,
			SelectorName: resInfo.ReferParam,
			SecondName:   resInfo.StructName,
		}, currentFile, handlerFnDecl.Body.List)
		if reqResInfo.StructName == "" {
			break
			//panic("reqResInfo.SelectorName is empty, " + bindReqName)
			//return nil, ""
		}
		schema, reqResFullName = a.getReqResSchema(reqResInfo, urlInfo)
	}
	urlInfo.ResIsArray = reqResInfo.IsArray
	urlInfo.ResSwagger = schema
	urlInfo.ResParam = reqResFullName
	a.urlMap[urlInfo.UniqueKey] = urlInfo
}

// 解析 handler 响应
// TODO: 解析 gin 的响应函数. c.JSON(), c.XML(), c.Data() 等
func parseHandlerResponse(fun *ast.SelectorExpr, handlerFnDecl *ast.FuncDecl, nn *ast.CallExpr, currentFile *ast.File) (resList []ResStruct) {
	resStruct := ResStruct{
		FuncName:  fun.Sel.Name,
		FuncValue: handlerFnDecl,
	}
	if nn.Args == nil {
		resStruct.Type = 1
		resList = append(resList, resStruct)
		return
	}

	switch argsT := nn.Args[0].(type) {
	case *ast.CompositeLit:
		switch specType := argsT.Type.(type) {
		// req :=  GoodCreateReq
		case *ast.Ident:
			resStruct.ReqResInfo = ReqResInfo{
				StructLocFile: currentFile,
				StructName:    specType.Name,
			}
			resStruct.Type = 2
			resList = append(resList, resStruct)

		// c.JSONOK(gin.H{})
		// c.JSONOK(dto.GoodRes{})
		case *ast.SelectorExpr:
			//locFile := a.packages.findPackagePathFromImports(specType.X.(*ast.Ident).String(), currentFile, false)
			//resList = append(resList, ResStruct{
			//	ReqResInfo: ReqResInfo{
			//		//StructLocFile: currentFile,
			//		StructName: specType.X.(*ast.Ident).String() + "." + specType.Sel.String(),
			//	},
			//	Type: 2,
			//})
			resStruct.ReqResInfo = ReqResInfo{
				StructLocFile: currentFile,
				StructName:    specType.X.(*ast.Ident).String() + "." + specType.Sel.String(),
			}
			resStruct.Type = 2
			resList = append(resList, resStruct)
		}

	case *ast.UnaryExpr:
		switch rhsTypeXType := argsT.X.(type) {
		case *ast.CompositeLit:
			switch specType := rhsTypeXType.Type.(type) {
			// req :=  GoodCreateReq
			case *ast.Ident:
				resStruct.ReqResInfo = ReqResInfo{
					StructLocFile: currentFile,
					StructName:    specType.Name,
				}
				resStruct.Type = 3
				resList = append(resList, resStruct)
				// c.JSONOK(&dto.GoodRes{})
			case *ast.SelectorExpr:
				resStruct.ReqResInfo = ReqResInfo{
					StructName: specType.X.(*ast.Ident).String() + "." + specType.Sel.String(),
				}
				resStruct.Type = 3
				resList = append(resList, resStruct)
			}
		}

	case *ast.Ident:
		// res := Struct{}
		// c.JSONOK(res)
		resParam := argsT.String()
		resStruct.Type = 4
		resStruct.ReferParam = resParam
		resList = append(resList, resStruct)

	case *ast.SelectorExpr:
		// 可能是赋值变量，也可能是某个函数变量
		resParam := argsT.X.(*ast.Ident).String()
		resStruct.ReqResInfo = ReqResInfo{
			StructName: argsT.Sel.String(),
		}
		resStruct.Type = 5
		resStruct.ReferParam = resParam
		resList = append(resList, resStruct)
	}

	return
}

// 解析请求参数绑定
func parseRequestParamBinding(handlerFnDecl *ast.FuncDecl, ctxName string) string {
	var bindReqName string
	// 因为会存在两种类型，所以直接使用ast.Inspect找到对应的 *ast.AssignStme
	// if err := c.Bind() xxxx
	// err := c.Bind()
	ast.Inspect(handlerFnDecl, func(n ast.Node) bool {
		if !handlerFnDecl.Name.IsExported() {
			return true
		}
		switch nn := n.(type) {
		case *ast.CallExpr:
			fun, flag := nn.Fun.(*ast.SelectorExpr)
			if !flag {
				return true
			}
			funcXIndent, flag := fun.X.(*ast.Ident)
			if !flag {
				return true
			}

			// 找到 c.Bind() 调用
			if funcXIndent.String() == ctxName && fun.Sel.Name == "Bind" {
				switch argType := nn.Args[0].(type) {
				case *ast.UnaryExpr:
					bindReqName = argType.X.(*ast.Ident).String()
				case *ast.Ident:
					bindReqName = argType.String()
				}
				return false
			}
			// TODO 解析注解 -请求参数类型
		}
		return true
	})
	return bindReqName
}

func (p *astParser) getReqResSchema(reqResInfo ReqResInfo, urlInfo UrlInfo) (schema *spec.Schema, reqAndResFullName string) {
	//if reqResInfo.StructLocFile == nil {
	//	panic("struct name: " + reqResInfo.StructName + ", loc file is nil")
	//}
	//todo

	if reqResInfo.FieldInfo != "" {
		fmt.Printf("reqResInfo.FieldValue--------------->"+"%+v\n", reqResInfo.FieldInfo)
		fmt.Printf("reqResInfo.FieldValue2--------------->"+"%+v\n", reqResInfo.StructLocFile.Name.String()+"."+reqResInfo.FieldValue)
		//reqAndResFullName = reqResInfo.StructLocFile.Name.String() + "." + reqResInfo.FieldValue
		reqAndResFullName = reqResInfo.FieldInfo
	} else {
		if reqResInfo.StructLocFile != nil {
			if strings.Contains(reqResInfo.StructName, ".") {
				reqAndResFullName = reqResInfo.StructName
				//if strings.HasPrefix(reqResInfo.StructName, urlInfo.ModuleName+".") {
				//	reqAndResFullName = strings.TrimPrefix(reqResInfo.StructName, urlInfo.ModuleName+".")
				//}
			} else {
				reqAndResFullName = reqResInfo.StructLocFile.Name.String() + "." + reqResInfo.StructName
			}
		} else {
			reqAndResFullName = reqResInfo.StructName
		}
	}

	var err error
	schema, err = p.getTypeSchema(reqResInfo.StructName, reqResInfo.FieldName, reqResInfo.StructLocFile, true)
	if err != nil {
		panic(err)
	}
	//// 说明这个类型就在这个函数里，直接使用type解析即可
	//if reqResInfo.ModName == "" {
	//	schemaInfo, err := p.getTypeSchema(reqResInfo.StructName, file, true)
	//	if err != nil {
	//		panic(err)
	//	}
	//	schema = schemaInfo
	//	reqAndResFullName = urlInfo.ModuleName + "." + reqResInfo.StructName
	//	// 说明在go其他包里
	//	// 先找到他引用的包
	//} else {
	//	importMapInfo := p.getImport(file)
	//	importInfo, flag := importMapInfo[reqResInfo.ModName]
	//	if !flag {
	//		panic("not find import info, modName: " + reqResInfo.ModName)
	//	}
	//	// 遍历去找这个type类型
	//	structFile, err := p.packages.findFileByRangePackages(importInfo.PackagePath, func(filename string, file *ast.File) (bool, error) {
	//		// 循环遍历
	//		var isContinueForeach = true
	//		ast.Inspect(file, func(n ast.Node) bool {
	//			if file.Name.String() == reqResInfo.ModName {
	//				switch nn := n.(type) {
	//				case *ast.GenDecl:
	//					// 定义的地方
	//					if nn.Tok == token.TYPE {
	//						info, flag := nn.Specs[0].(*ast.TypeSpec)
	//						if !flag {
	//							return true
	//						}
	//						if info.Name.String() == reqResInfo.StructName {
	//							// 找到了，不需要再循环
	//							isContinueForeach = false
	//							return false
	//						}
	//					}
	//				}
	//			}
	//			return true
	//		})
	//
	//		return isContinueForeach, nil
	//	})
	//
	//	if err != nil {
	//		panic(err)
	//	}
	//	schemaInfo, err := p.getTypeSchema(reqResInfo.StructName, structFile, true)
	//	if err != nil {
	//		fmt.Printf("err--------------->"+"%+v\n", err)
	//	}
	//	schema = schemaInfo
	//	reqAndResFullName = reqResInfo.ModName + "." + reqResInfo.StructName
	//}
	return
}

func (p *astParser) GetData() []UrlInfo {
	output := make([]UrlInfo, 0)
	for _, value := range p.sortedUrl {
		output = append(output, p.urlMap[value.UniqueKey])
	}
	return output
}

func (p *astParser) GetDefinitions() spec.Definitions {
	return p.swagger.Definitions
}

func (p *astParser) getTypeSchema(typeName string, fieldName string, file *ast.File, ref bool) (*spec.Schema, error) {
	if IsGolangPrimitiveType(typeName) {
		return PrimitiveSchema(TransToValidSchemeType(typeName)), nil
	}

	typeSpecDef := p.packages.FindTypeSpec(typeName, file, p.ParseAllDependency)
	if typeSpecDef == nil {
		return nil, fmt.Errorf("cannot find type definition: %s", typeName)
	}

	//fmt.Printf("typeSpecDef--------------->"+"%+v\n", typeSpecDef)
	schema, ok := p.parsedSchemas[typeSpecDef]
	if !ok {
		var err error
		// 解析结构体数据
		// 如果field name为空，解析第一层
		// 如果field name有数据，返回这个结构体
		schema, err = p.ParseDefinition(typeSpecDef, fieldName)
		if err != nil {
			if err == ErrRecursiveParseStruct && ref {
				return p.getRefTypeSchema(typeSpecDef, schema), nil
			}

			return nil, err
		}
	}
	if fieldName != "" {
		structInfo := typeSpecDef.TypeSpec.Type.(*ast.StructType)
		for _, field := range structInfo.Fields.List {
			if field.Names[0].String() != fieldName {
				continue
			}
			typeName, err := getFieldType(field.Type)
			if err == nil {
				// named type
				fmt.Printf("named type--------------->"+"%+v\n", typeName)
				infoSchema, err := p.getTypeSchema(typeName, "", file, true)
				if err != nil {
					panic(err)
				}

				return infoSchema, nil
			} else {
				// unnamed type
				infoSchema, err := p.parseTypeExpr(file, field.Type, false)
				if err != nil {
					panic(err)
				}
				return infoSchema, nil
			}
		}
		panic("not found field " + fieldName)
	}

	// 如果ref为true，同时是对象类型
	if ref && len(schema.Schema.Type) > 0 && schema.Schema.Type[0] == OBJECT {
		return p.getRefTypeSchema(typeSpecDef, schema), nil
	}

	if ref && len(schema.Schema.Type) > 0 && schema.Schema.Type[0] == ARRAY {
		return p.getRefTypeSchema(typeSpecDef, schema), nil
	}

	return schema.Schema, nil
}

// ParseDefinition parses given type spec that corresponds to the type under
// given name and package, and populates swagger schema definitions registry
// with a schema for the given type
func (p *astParser) ParseDefinition(typeSpecDef *TypeSpecDef, fieldName string) (*Schema, error) {
	typeName := typeSpecDef.FullName()
	refTypeName := TypeDocName(typeName, typeSpecDef.TypeSpec)
	schema, ok := p.parsedSchemas[typeSpecDef]
	if ok {
		elog.Debugf("Skipping '%s', already parsed.", typeName)
		return schema, nil
	}

	if p.isInStructStack(typeSpecDef) {
		//p.debug.Printf("Skipping '%s', recursion detected.", typeName)

		return &Schema{
				Name:    refTypeName,
				PkgPath: typeSpecDef.PkgPath,
				Schema:  PrimitiveSchema(OBJECT),
			},
			ErrRecursiveParseStruct
	}
	p.structStack = append(p.structStack, typeSpecDef)

	//p.debug.Printf("Generating %s", typeName)
	fmt.Println(fmt.Sprintf("Generating %s", typeName))
	definition, err := p.parseTypeExpr(typeSpecDef.File, typeSpecDef.TypeSpec.Type, false)
	if err != nil {
		return nil, err
	}

	if definition.Description == "" {
		fillDefinitionDescription(definition, typeSpecDef.File, typeSpecDef)
	}

	s := Schema{
		Name:    refTypeName,
		PkgPath: typeSpecDef.PkgPath,
		Schema:  definition,
	}
	p.parsedSchemas[typeSpecDef] = &s

	// update an empty schema as a result of recursion
	s2, ok := p.outputSchemas[typeSpecDef]
	if ok {
		p.swagger.Definitions[s2.Name] = *definition
	}

	return &s, nil
}

// fillDefinitionDescription additionally fills fields in definition (spec.Schema)
// TODO: If .go file contains many types, it may work for a long time
func fillDefinitionDescription(definition *spec.Schema, file *ast.File, typeSpecDef *TypeSpecDef) {
	for _, astDeclaration := range file.Decls {
		generalDeclaration, ok := astDeclaration.(*ast.GenDecl)
		if !ok || generalDeclaration.Tok != token.TYPE {
			continue
		}

		for _, astSpec := range generalDeclaration.Specs {
			typeSpec, ok := astSpec.(*ast.TypeSpec)
			if !ok || typeSpec != typeSpecDef.TypeSpec {
				continue
			}

			definition.Description =
				extractDeclarationDescription(typeSpec.Doc, typeSpec.Comment, generalDeclaration.Doc)
		}
	}
}

// extractDeclarationDescription gets first description
// from attribute descriptionAttr in commentGroups (ast.CommentGroup)
func extractDeclarationDescription(commentGroups ...*ast.CommentGroup) string {
	var description string

	for _, commentGroup := range commentGroups {
		if commentGroup == nil {
			continue
		}

		isHandlingDescription := false
		for _, comment := range commentGroup.List {
			commentText := strings.TrimSpace(strings.TrimLeft(comment.Text, "/"))
			attribute := strings.Split(commentText, " ")[0]
			if strings.ToLower(attribute) != descriptionAttr {
				if !isHandlingDescription {
					continue
				}
				break
			}

			isHandlingDescription = true
			description += " " + strings.TrimSpace(commentText[len(attribute):])
		}
	}

	return strings.TrimLeft(description, " ")
}

// parseTypeExpr parses given type expression that corresponds to the type under
// given name and package, and returns swagger schema for it.
func (p *astParser) parseTypeExpr(file *ast.File, typeExpr ast.Expr, ref bool) (*spec.Schema, error) {
	switch expr := typeExpr.(type) {
	// type Foo interface{}
	case *ast.InterfaceType:
		return &spec.Schema{}, nil

	// type Foo struct {...}
	case *ast.StructType:
		return p.parseStruct(file, expr.Fields)

	// type Foo Baz
	case *ast.Ident:
		return p.getTypeSchema(expr.Name, "", file, ref)

	// type Foo *Baz
	case *ast.StarExpr:
		return p.parseTypeExpr(file, expr.X, ref)

	// type Foo pkg.Bar
	case *ast.SelectorExpr:
		if xIdent, ok := expr.X.(*ast.Ident); ok {
			return p.getTypeSchema(fullTypeName(xIdent.Name, expr.Sel.Name), "", file, ref)
		}
	// type Foo []Baz
	case *ast.ArrayType:
		itemSchema, err := p.parseTypeExpr(file, expr.Elt, true)
		if err != nil {
			return nil, err
		}

		return spec.ArrayProperty(itemSchema), nil
	// type Foo map[string]Bar
	case *ast.MapType:
		if _, ok := expr.Value.(*ast.InterfaceType); ok {
			return spec.MapProperty(nil), nil
		}
		schema, err := p.parseTypeExpr(file, expr.Value, true)
		if err != nil {
			return nil, err
		}

		return spec.MapProperty(schema), nil

	case *ast.FuncType:
		return nil, ErrFuncTypeField
	// ...
	default:
		panic(fmt.Sprintf("Type definition of type '%T' is not supported yet. Using 'object' instead.\n", typeExpr))
	}

	return PrimitiveSchema(OBJECT), nil
}

func (p *astParser) getRefTypeSchema(typeSpecDef *TypeSpecDef, schema *Schema) *spec.Schema {
	_, ok := p.outputSchemas[typeSpecDef]
	if !ok {
		existSchema, ok := p.existSchemaNames[schema.Name]
		if ok {
			// store the first one to be renamed after parsing over
			_, ok = p.toBeRenamedSchemas[existSchema.Name]
			if !ok {
				p.toBeRenamedSchemas[existSchema.Name] = existSchema.PkgPath
			}
			// rename not the first one
			schema.Name = p.renameSchema(schema.Name, schema.PkgPath)
		} else {
			p.existSchemaNames[schema.Name] = schema
		}
		p.swagger.Definitions[schema.Name] = spec.Schema{}

		if schema.Schema != nil {
			p.swagger.Definitions[schema.Name] = *schema.Schema
		}

		p.outputSchemas[typeSpecDef] = schema
	}

	refSchema := RefSchema(schema.Name)
	// store every URL
	p.toBeRenamedRefURLs = append(p.toBeRenamedRefURLs, refSchema.Ref.GetURL())

	return refSchema
}

// parseStruct 解析结构体都会到这里来
func (p *astParser) parseStruct(file *ast.File, fields *ast.FieldList) (*spec.Schema, error) {
	required := make([]string, 0)
	properties := make(map[string]spec.Schema)
	for _, field := range fields.List {
		fieldProps, requiredFromAnon, err := p.parseStructField(file, field)
		if err != nil {
			if err == ErrFuncTypeField || err == ErrSkippedField {
				continue
			}

			return nil, err
		}
		if len(fieldProps) == 0 {
			continue
		}
		required = append(required, requiredFromAnon...)
		for k, v := range fieldProps {
			properties[k] = v
		}
	}

	sort.Strings(required)

	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:       []string{OBJECT},
			Properties: properties,
			Required:   required,
		},
	}, nil
}

/*
 field.Names--------------->[Title]
 field.Names--------------->[SubTitle]
 field.Names--------------->[Cover]
 field.Names--------------->[Arr]
 field.Names--------------->[Co]
*/
func (p *astParser) parseStructField(file *ast.File, field *ast.Field) (map[string]spec.Schema, []string, error) {
	if field.Names == nil {
		if field.Tag != nil {
			skip, ok := reflect.StructTag(strings.ReplaceAll(field.Tag.Value, "`", "")).Lookup("swaggerignore")
			if ok && strings.EqualFold(skip, "true") {
				return nil, nil, nil
			}
		}

		typeName, err := getFieldType(field.Type)
		if err != nil {
			return nil, nil, err
		}
		schema, err := p.getTypeSchema(typeName, "", file, false)
		if err != nil {
			return nil, nil, err
		}
		if len(schema.Type) > 0 && schema.Type[0] == OBJECT {
			if len(schema.Properties) == 0 {
				return nil, nil, nil
			}

			properties := map[string]spec.Schema{}
			for k, v := range schema.Properties {
				properties[k] = v
			}

			return properties, schema.SchemaProps.Required, nil
		}

		// for alias type of non-struct types ,such as array,map, etc. ignore field tag.
		return map[string]spec.Schema{typeName: *schema}, nil, nil
	}

	ps := p.fieldParserFactory(p, field)

	ok, err := ps.ShouldSkip()
	if err != nil {
		return nil, nil, err
	}
	if ok {
		return nil, nil, nil
	}

	fieldName, err := ps.FieldName()
	if err != nil {
		return nil, nil, err
	}

	schema, err := ps.CustomSchema()
	if err != nil {
		return nil, nil, err
	}
	if schema == nil {
		typeName, err := getFieldType(field.Type)
		if err == nil {
			// named type
			schema, err = p.getTypeSchema(typeName, "", file, true)
		} else {
			// unnamed type
			schema, err = p.parseTypeExpr(file, field.Type, false)
		}
		if err != nil {
			return nil, nil, err
		}
	}

	err = ps.ComplementSchema(schema)
	if err != nil {
		return nil, nil, err
	}

	var tagRequired []string
	required, err := ps.IsRequired()
	if err != nil {
		return nil, nil, err
	}
	if required {
		tagRequired = append(tagRequired, fieldName)
	}

	return map[string]spec.Schema{fieldName: *schema}, tagRequired, nil
}

func getFieldType(field ast.Expr) (string, error) {
	switch fieldType := field.(type) {
	case *ast.Ident:
		return fieldType.Name, nil
	case *ast.SelectorExpr:
		packageName, err := getFieldType(fieldType.X)
		if err != nil {
			return "", err
		}
		return fullTypeName(packageName, fieldType.Sel.Name), nil
	case *ast.StarExpr:
		fullName, err := getFieldType(fieldType.X)
		if err != nil {
			return "", err
		}
		return fullName, nil
	default:
		return "", fmt.Errorf("unknown field type %#v", field)
	}
}

func getFieldTypeX(field ast.Expr) string {
	switch fieldType := field.(type) {
	case *ast.Ident:
		return fieldType.Name
	case *ast.SelectorExpr:
		packageName := getFieldTypeX(fieldType.X)
		return fullTypeName(packageName, fieldType.Sel.Name)
	case *ast.StarExpr:
		fullName := getFieldTypeX(fieldType.X)
		return fullName
	default:
		panic(fmt.Errorf("unknown field type %#v", field).Error())
	}
}

func (p *astParser) renameSchema(name, pkgPath string) string {
	parts := strings.Split(name, ".")
	name = fullTypeName(pkgPath, parts[len(parts)-1])
	name = strings.ReplaceAll(name, "/", "_")

	return name
}

func (p *astParser) isInStructStack(typeSpecDef *TypeSpecDef) bool {
	for _, specDef := range p.structStack {
		if typeSpecDef == specDef {
			return true
		}
	}

	return false
}

//
//func (a *astParser) initReadContent() error {
//	//if a.userOption.ScaffoldDSLContent == "" {
//	//	if a.userOption.RootFile == "" {
//	//		return fmt.Errorf("content and file is empty")
//	//	}
//	//	//contentByte, err := ioutil.ReadFile(a.userOption.RootFile)
//	//	//if err != nil {
//	//	//	panic("initReadContent: " + err.Error())
//	//	//}
//	//	//a.userOption.ScaffoldDSLContent = string(contentByte)
//	//	//if a.userOption.ScaffoldDSLContent == "" {
//	//	//	return fmt.Errorf("内容不能为空")
//	//	//}
//	//}
//	//a.readContent = a.userOption.ScaffoldDSLContent
//	return nil
//}

/**
*ast.File {
     1  .  Doc: nil
     2  .  Package: foo:1:1
     3  .  Name: *ast.Ident {
     4  .  .  NamePos: foo:1:9
     5  .  .  Name: "testdata"
     6  .  .  Obj: nil
     7  .  }
     8  .  Decls: []ast.Decl (len = 2) {
     9  .  .  0: *ast.GenDecl {
    10  .  .  .  Doc: nil
    11  .  .  .  TokPos: foo:3:1
    12  .  .  .  Tok: import
    13  .  .  .  Lparen: foo:3:8
    14  .  .  .  Specs: []ast.Spec (len = 2) {
    15  .  .  .  .  0: *ast.ImportSpec {
    16  .  .  .  .  .  Doc: nil
    17  .  .  .  .  .  Name: nil
    18  .  .  .  .  .  Path: *ast.BasicLit {
    19  .  .  .  .  .  .  ValuePos: foo:4:2
    20  .  .  .  .  .  .  Kind: STRING
    21  .  .  .  .  .  .  Value: "\"net/http\""
    22  .  .  .  .  .  }
    23  .  .  .  .  .  Comment: nil
    24  .  .  .  .  .  EndPos: -
    25  .  .  .  .  }
    26  .  .  .  .  1: *ast.ImportSpec {
    27  .  .  .  .  .  Doc: nil
    28  .  .  .  .  .  Name: nil
    29  .  .  .  .  .  Path: *ast.BasicLit {
    30  .  .  .  .  .  .  ValuePos: foo:5:2
    31  .  .  .  .  .  .  Kind: STRING
    32  .  .  .  .  .  .  Value: "\"strings\""
    33  .  .  .  .  .  }
    34  .  .  .  .  .  Comment: nil
    35  .  .  .  .  .  EndPos: -
    36  .  .  .  .  }
    37  .  .  .  }
    38  .  .  .  Rparen: foo:8:1
    39  .  .  }
    40  .  .  1: *ast.FuncDecl {
    41  .  .  .  Doc: nil
    42  .  .  .  Recv: nil
    43  .  .  .  Name: *ast.Ident {
    44  .  .  .  .  NamePos: foo:12:6
    45  .  .  .  .  Name: "ServeHttp"
    46  .  .  .  .  Obj: *ast.Object {
    47  .  .  .  .  .  Kind: func
    48  .  .  .  .  .  Name: "ServeHttp"
    49  .  .  .  .  .  Decl: *(obj @ 40)
    50  .  .  .  .  .  Data: nil
    51  .  .  .  .  .  Type: nil
    52  .  .  .  .  }
    53  .  .  .  }
    54  .  .  .  Type: *ast.FuncType {
    55  .  .  .  .  Func: foo:12:1
    56  .  .  .  .  Params: *ast.FieldList {
    57  .  .  .  .  .  Opening: foo:12:15
    58  .  .  .  .  .  List: nil
    59  .  .  .  .  .  Closing: foo:12:16
    60  .  .  .  .  }
    61  .  .  .  .  Results: *ast.FieldList {
    62  .  .  .  .  .  Opening: -
    63  .  .  .  .  .  List: []*ast.Field (len = 1) {
    64  .  .  .  .  .  .  0: *ast.Field {
    65  .  .  .  .  .  .  .  Doc: nil
    66  .  .  .  .  .  .  .  Names: nil
    67  .  .  .  .  .  .  .  Type: *ast.StarExpr {
    68  .  .  .  .  .  .  .  .  Star: foo:12:18
    69  .  .  .  .  .  .  .  .  X: *ast.SelectorExpr {
    70  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
    71  .  .  .  .  .  .  .  .  .  .  NamePos: foo:12:19
    72  .  .  .  .  .  .  .  .  .  .  Name: "egin"
    73  .  .  .  .  .  .  .  .  .  .  Obj: nil
    74  .  .  .  .  .  .  .  .  .  }
    75  .  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
    76  .  .  .  .  .  .  .  .  .  .  NamePos: foo:12:24
    77  .  .  .  .  .  .  .  .  .  .  Name: "Component"
    78  .  .  .  .  .  .  .  .  .  .  Obj: nil
    79  .  .  .  .  .  .  .  .  .  }
    80  .  .  .  .  .  .  .  .  }
    81  .  .  .  .  .  .  .  }
    82  .  .  .  .  .  .  .  Tag: nil
    83  .  .  .  .  .  .  .  Comment: nil
    84  .  .  .  .  .  .  }
    85  .  .  .  .  .  }
    86  .  .  .  .  .  Closing: -
    87  .  .  .  .  }
    88  .  .  .  }
    89  .  .  .  Body: *ast.BlockStmt {
    90  .  .  .  .  Lbrace: foo:12:34
    91  .  .  .  .  List: []ast.Stmt (len = 3) {
    92  .  .  .  .  .  0: *ast.AssignStmt {
    93  .  .  .  .  .  .  Lhs: []ast.Expr (len = 1) {
    94  .  .  .  .  .  .  .  0: *ast.Ident {
    95  .  .  .  .  .  .  .  .  NamePos: foo:13:2
    96  .  .  .  .  .  .  .  .  Name: "r"
    97  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    98  .  .  .  .  .  .  .  .  .  Kind: var
    99  .  .  .  .  .  .  .  .  .  Name: "r"
   100  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 92)
   101  .  .  .  .  .  .  .  .  .  Data: nil
   102  .  .  .  .  .  .  .  .  .  Type: nil
   103  .  .  .  .  .  .  .  .  }
   104  .  .  .  .  .  .  .  }
   105  .  .  .  .  .  .  }
   106  .  .  .  .  .  .  TokPos: foo:13:4
   107  .  .  .  .  .  .  Tok: :=
   108  .  .  .  .  .  .  Rhs: []ast.Expr (len = 1) {
   109  .  .  .  .  .  .  .  0: *ast.SelectorExpr {
   110  .  .  .  .  .  .  .  .  X: *ast.Ident {
   111  .  .  .  .  .  .  .  .  .  NamePos: foo:13:7
   112  .  .  .  .  .  .  .  .  .  Name: "invoker"
   113  .  .  .  .  .  .  .  .  .  Obj: nil
   114  .  .  .  .  .  .  .  .  }
   115  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
   116  .  .  .  .  .  .  .  .  .  NamePos: foo:13:15
   117  .  .  .  .  .  .  .  .  .  Name: "Gin"
   118  .  .  .  .  .  .  .  .  .  Obj: nil
   119  .  .  .  .  .  .  .  .  }
   120  .  .  .  .  .  .  .  }
   121  .  .  .  .  .  .  }
   122  .  .  .  .  .  }
   123  .  .  .  .  .  1: *ast.ExprStmt {
   124  .  .  .  .  .  .  X: *ast.CallExpr {
   125  .  .  .  .  .  .  .  Fun: *ast.SelectorExpr {
   126  .  .  .  .  .  .  .  .  X: *ast.Ident {
   127  .  .  .  .  .  .  .  .  .  NamePos: foo:16:2
   128  .  .  .  .  .  .  .  .  .  Name: "r"
   129  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 97)
   130  .  .  .  .  .  .  .  .  }
   131  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
   132  .  .  .  .  .  .  .  .  .  NamePos: foo:16:4
   133  .  .  .  .  .  .  .  .  .  Name: "POST"
   134  .  .  .  .  .  .  .  .  .  Obj: nil
   135  .  .  .  .  .  .  .  .  }
   136  .  .  .  .  .  .  .  }
   137  .  .  .  .  .  .  .  Lparen: foo:16:8
   138  .  .  .  .  .  .  .  Args: []ast.Expr (len = 2) {
   139  .  .  .  .  .  .  .  .  0: *ast.BasicLit {
   140  .  .  .  .  .  .  .  .  .  ValuePos: foo:16:9
   141  .  .  .  .  .  .  .  .  .  Kind: STRING
   142  .  .  .  .  .  .  .  .  .  Value: "\"/api/callbacks/wechat\""
   143  .  .  .  .  .  .  .  .  }
   144  .  .  .  .  .  .  .  .  1: *ast.CallExpr {
   145  .  .  .  .  .  .  .  .  .  Fun: *ast.Ident {
   146  .  .  .  .  .  .  .  .  .  .  NamePos: foo:16:34
   147  .  .  .  .  .  .  .  .  .  .  Name: "H"
   148  .  .  .  .  .  .  .  .  .  .  Obj: nil
   149  .  .  .  .  .  .  .  .  .  }
   150  .  .  .  .  .  .  .  .  .  Lparen: foo:16:35
   151  .  .  .  .  .  .  .  .  .  Args: []ast.Expr (len = 1) {
   152  .  .  .  .  .  .  .  .  .  .  0: *ast.SelectorExpr {
   153  .  .  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
   154  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: foo:16:36
   155  .  .  .  .  .  .  .  .  .  .  .  .  Name: "shop"
   156  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   157  .  .  .  .  .  .  .  .  .  .  .  }
   158  .  .  .  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
   159  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: foo:16:41
   160  .  .  .  .  .  .  .  .  .  .  .  .  Name: "CallbackWechat"
   161  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   162  .  .  .  .  .  .  .  .  .  .  .  }
   163  .  .  .  .  .  .  .  .  .  .  }
   164  .  .  .  .  .  .  .  .  .  }
   165  .  .  .  .  .  .  .  .  .  Ellipsis: -
   166  .  .  .  .  .  .  .  .  .  Rparen: foo:16:55
   167  .  .  .  .  .  .  .  .  }
   168  .  .  .  .  .  .  .  }
   169  .  .  .  .  .  .  .  Ellipsis: -
   170  .  .  .  .  .  .  .  Rparen: foo:16:56
   171  .  .  .  .  .  .  }
   172  .  .  .  .  .  }
   173  .  .  .  .  .  2: *ast.ReturnStmt {
   174  .  .  .  .  .  .  Return: foo:18:2
   175  .  .  .  .  .  .  Results: []ast.Expr (len = 1) {
   176  .  .  .  .  .  .  .  0: *ast.Ident {
   177  .  .  .  .  .  .  .  .  NamePos: foo:18:9
   178  .  .  .  .  .  .  .  .  Name: "r"
   179  .  .  .  .  .  .  .  .  Obj: *(obj @ 97)
   180  .  .  .  .  .  .  .  }
   181  .  .  .  .  .  .  }
   182  .  .  .  .  .  }
   183  .  .  .  .  }
   184  .  .  .  .  Rbrace: foo:19:1
   185  .  .  .  }
   186  .  .  }
   187  .  }
   188  .  Scope: *ast.Scope {
   189  .  .  Outer: nil
   190  .  .  Objects: map[string]*ast.Object (len = 1) {
   191  .  .  .  "ServeHttp": *(obj @ 46)
   192  .  .  }
   193  .  }
   194  .  Imports: []*ast.ImportSpec (len = 2) {
   195  .  .  0: *(obj @ 15)
   196  .  .  1: *(obj @ 26)
   197  .  }
   198  .  Unresolved: []*ast.Ident (len = 4) {
   199  .  .  0: *(obj @ 70)
   200  .  .  1: *(obj @ 110)
   201  .  .  2: *(obj @ 145)
   202  .  .  3: *(obj @ 153)
   203  .  }
   204  .  Comments: []*ast.CommentGroup (len = 2) {
   205  .  .  0: *ast.CommentGroup {
   206  .  .  .  List: []*ast.Comment (len = 1) {
   207  .  .  .  .  0: *ast.Comment {
   208  .  .  .  .  .  Slash: foo:10:1
   209  .  .  .  .  .  Text: "//var H = bffcore.Handle"
   210  .  .  .  .  }
   211  .  .  .  }
   212  .  .  }
   213  .  .  1: *ast.CommentGroup {
   214  .  .  .  List: []*ast.Comment (len = 1) {
   215  .  .  .  .  0: *ast.Comment {
   216  .  .  .  .  .  Slash: foo:16:58
   217  .  .  .  .  .  Text: "// 微信通知回调接口"
   218  .  .  .  .  }
   219  .  .  .  }
   220  .  .  }
   221  .  }
   222  }
*/

func (p *astParser) findEgoComponentFile() (*ast.File, error) {
	return p.packages.findFileByRangeFiles(func(filename string, file *ast.File) (bool, error) {
		info, _ := p.findEginRouter(file)
		if info != nil {
			return false, nil
		}
		return true, nil
	})
}

func (p *astParser) parserStruct() error {

	f, err := p.findEgoComponentFile()
	if err != nil {
		panic(err)
	}

	funcValue, orderNum := p.findEginRouter(f)
	p.findRouterParam(funcValue, orderNum)
	p.findImport(f)
	//p.findGroup()
	for _, stmtValue := range funcValue.Body.List {
		p.dealGroup(stmtValue)
		p.findUrl(stmtValue)
		//stmt, flag := stmtValue.(*ast.ExprStmt)
		//if !flag {
		//	continue
		//}
		//stmtCall, flag := stmt.X.(*ast.CallExpr)
		//if !flag {
		//	continue
		//}
		//funcSelector, flag := stmtCall.Fun.(*ast.SelectorExpr)
		//if !flag {
		//	continue
		//}
		//routerIndent, flag := funcSelector.X.(*ast.Ident)
		//if !flag {
		//	continue
		//}
		//
		//// 找到变量，在找
		//if routerIndent.Name == p.routerParam {
		//	fmt.Println("111111")
		//}
		//.Fun.(*ast.SelectorExpr).Sel.Name = "POST"
		//stmt.X.(*ast.CallExpr).Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name = "H"
	}

	scope := f.Scope
	if scope == nil {
		return errors.New("struct nil")
	}
	//objects := scope.Objects
	//fmt.Printf("objects--------------->"+"%+v\n", objects["ServeHttp"])

	//structs := make([]*SpecType, 0)
	//for structName, obj := range objects {
	//	st, err := a.parseObject(structName, obj)
	//	if err != nil {
	//		return err
	//	}
	//	structs = append(structs, st)
	//}
	//sort.Slice(structs, func(i, j int) bool {
	//	return structs[i].Name < structs[j].Name
	//})
	//
	//resp := make([]SpecType, 0)
	//for _, item := range structs {
	//	resp = append(resp, *item)
	//}
	//a.modelArr = resp
	return nil
}

/**
如何确定一个ego的gin router
在返回值中，只要有一个参数是*egin.Component，就是我们的router函数
然后我们在从return中找到对应顺序的变量名
在然后通过解析这个函数里数据，如果是这个变量名，并且是函数调用方法，使用的Group、GET、POST等名称，那么就是我们的方法
再然后通过import找到对应的文件路径，解析他的dto
*/
func (p *astParser) findEginRouter(f *ast.File) (funcValue *ast.FuncDecl, outputOrderNum int) {
	for _, value := range f.Decls {
		ast.Inspect(value, func(n ast.Node) bool {
			switch nn := n.(type) {
			case *ast.FuncDecl:
				// HttpServer
				// valueData.Name
				// 这里放着参数、返回值
				// orderNum是第几个顺序
				// 因为是多个返回值，所以我们需要遍历*ast.Field，我们认为*egin.Component，就是我们的所有router项目
				if nn.Type.Results == nil {
					return true
				}
				for orderNum, fieldValue := range nn.Type.Results.List {
					resultFieldPointer, flag := fieldValue.Type.(*ast.StarExpr)
					if !flag {
						continue
					}
					selectorExpr, flag := resultFieldPointer.X.(*ast.SelectorExpr)
					if !flag {
						continue
					}
					name := selectorExpr.X.(*ast.Ident).Name
					if name == "egin" {
						// 这个才是函数router
						if selectorExpr.Sel.Name == "Component" {
							p.routerFunc = nn
							funcValue = nn
							outputOrderNum = orderNum
							return false
						}
					}
				}
			}
			return true
		})
	}
	return
}

// 找到变量名
func (p *astParser) findRouterParam(decl *ast.FuncDecl, orderNum int) {
	for _, stmtValue := range decl.Body.List {
		// 找到返回值
		stmt, flag := stmtValue.(*ast.ReturnStmt)
		if !flag {
			continue
		}
		p.routerParam = stmt.Results[orderNum].(*ast.Ident).Name
		// 因为只有一个返回值直接return
		// 后续可以通过
		return
	}
	panic("no router param")
	return
}

// 找到变量名
func (p *astParser) findImport(f *ast.File) {
	p.importMap = p.getImport(f)
}

// 找到变量名
func (p *astParser) getImport(f *ast.File) map[string]ImportInfo {
	output := make(map[string]ImportInfo)
	//  f.Decls
	for _, value := range f.Decls {
		// *ast.FuncDecl
		// 拿到函数
		valueData, flag := value.(*ast.GenDecl)
		if !flag {
			continue
		}
		if valueData.Tok != token.IMPORT {
			continue
		}

		for _, spec := range valueData.Specs {
			importSpec, flag := spec.(*ast.ImportSpec)
			if !flag {
				continue
			}
			importPath := trimPath(importSpec.Path.Value)
			importPath = strings.Trim(importPath, p.GoMod+"/")
			// 如果不等于nil，说明import 有别名
			if importSpec.Name != nil {
				output[importSpec.Name.String()] = ImportInfo{
					KeyName:     importSpec.Name.String(),
					Alias:       importSpec.Name.String(),
					PackagePath: importPath,
				}
			} else {
				arr := strings.Split(importPath, "/")
				keyName := arr[len(arr)-1]
				output[keyName] = ImportInfo{
					KeyName:     keyName,
					Alias:       "",
					PackagePath: importPath,
				}
			}

		}
	}
	return output
}

// 找到group
func (p *astParser) findGroup() {
	for _, stmtValue := range p.routerFunc.Body.List {
		stmt, flag := stmtValue.(*ast.AssignStmt)
		if !flag {
			continue
		}

		// g := r.Group()
		if !(stmt.Tok.String() == ":=") {
			continue
		}

		assignValue := stmt.Rhs[0]
		stmtCall, flag := assignValue.(*ast.CallExpr)
		if !flag {
			continue
		}

		funcSelector, flag := stmtCall.Fun.(*ast.SelectorExpr)
		if !flag {
			continue
		}
		// r
		routerIndent, flag := funcSelector.X.(*ast.Ident)
		if !flag {
			continue
		}
		// 说明是router参数的名称，说明是分组
		if routerIndent.Name == p.routerParam && funcSelector.Sel.Name == "Group" {
			keyValue := stmt.Lhs[0].(*ast.Ident).Name
			p.groupMap[keyValue] = GroupInfo{
				keyValue,
				stmtCall.Args[0].(*ast.BasicLit).Value,
			}
		}
		return
	}
	return
}

func trimPath(path string) string {
	return strings.Trim(path, `"`)
}

/**
1. g := r.Group()
2. g = g.Group()
*/

func (p *astParser) dealGroup(stmtValue ast.Stmt) {
	stmt, flag := stmtValue.(*ast.AssignStmt)
	if !flag {
		return
	}

	// g := r.Group()
	if !(stmt.Tok.String() == ":=") && !(stmt.Tok.String() == "=") {
		return
	}

	assignValue := stmt.Rhs[0]
	stmtCall, flag := assignValue.(*ast.CallExpr)
	if !flag {
		return
	}

	funcSelector, flag := stmtCall.Fun.(*ast.SelectorExpr)
	if !flag {
		return
	}
	// r
	routerIndent, flag := funcSelector.X.(*ast.Ident)
	if !flag {
		return
	}

	// 去掉引号

	if stmt.Tok.String() == ":=" {
		// 说明是router参数的名称，说明是分组
		if routerIndent.Name == p.routerParam && funcSelector.Sel.Name == "Group" {
			keyValue := stmt.Lhs[0].(*ast.Ident).Name
			p.groupMap[keyValue] = GroupInfo{
				keyValue,
				strings.Trim(stmtCall.Args[0].(*ast.BasicLit).Value, `"`),
			}
			return
		}
	}

	if stmt.Tok.String() == "=" {
		// 说明是之前旧的router参数的名称，说明是分组
		groupInfo, flag := p.groupMap[routerIndent.Name]
		if flag {
			if funcSelector.Sel.Name == "Group" {
				keyValue := stmt.Lhs[0].(*ast.Ident).Name
				p.groupMap[keyValue] = GroupInfo{
					keyValue,
					groupInfo.Prefix + strings.Trim(stmtCall.Args[0].(*ast.BasicLit).Value, `"`),
				}
				return
			}
		}
	}
}

// 找到group
func (p *astParser) findUrl(stmtValue ast.Stmt) {
	stmt, flag := stmtValue.(*ast.ExprStmt)
	if !flag {
		return
	}
	callExpr, flag := stmt.X.(*ast.CallExpr)
	if !flag {
		return
	}

	funcSelector, flag := callExpr.Fun.(*ast.SelectorExpr)
	if !flag {
		return
	}
	// r
	routerIndent, flag := funcSelector.X.(*ast.Ident)
	if !flag {
		return
	}

	// 如果是根目录
	if routerIndent.Name == p.routerParam {
		p.getUrl(funcSelector.Sel.Name, "", callExpr)
		return
	}

	groupInfo, flag := p.groupMap[routerIndent.Name]
	if flag {
		p.getUrl(funcSelector.Sel.Name, groupInfo.Prefix, callExpr)
		return
	}
	return
}

func (p *astParser) getUrl(methodName string, prefix string, callExpr *ast.CallExpr) {
	if lo.Contains([]string{"POST", "GET", "DELETE", "PUT", "PATCH", "OPTION"}, methodName) {
		length := len(callExpr.Args)
		if length < 2 {
			panic("get url length < 2")
		}
		urlKey, flag := callExpr.Args[0].(*ast.BasicLit)
		if !flag {
			panic("get url type fail")
		}

		urlPath := strings.Trim(urlKey.Value, `"`)
		fullPath := prefix + urlPath
		funcInfo, flag := getFuncType(callExpr.Args[length-1])
		if !flag {
			return
		}
		urlInfo := UrlInfo{
			UniqueKey:  methodName + "." + fullPath,
			Method:     methodName,
			Prefix:     prefix,
			FullPath:   fullPath,
			ModuleName: funcInfo.ModuleName,
			FuncName:   funcInfo.FuncName,
		}

		// 寻找他的调用文件的module path
		importInfo, flag := p.importMap[funcInfo.ModuleName]
		if !flag {
			panic("cant get import info, module name: " + funcInfo.ModuleName)
		}
		urlInfo.PackagePath = importInfo.PackagePath

		_, flag = p.urlMap[urlInfo.UniqueKey]
		if flag {
			panic("register duplicate, full path: " + urlInfo.UniqueKey)
		}
		p.urlMap[urlInfo.UniqueKey] = urlInfo
		p.sortedUrl = append(p.sortedUrl, urlInfo)
		return
	}
}

type CallApiFunc struct {
	ModuleName string
	FuncName   string
}

func getFuncType(expr ast.Expr) (CallApiFunc, bool) {
	switch v := expr.(type) {
	case *ast.SelectorExpr:
		x := v.X
		xIdent, ok := x.(*ast.Ident)
		// todo invoker.Test.Create 这种函数的func暂时不好解析
		if !ok {
			return CallApiFunc{}, false
		}
		return CallApiFunc{
			ModuleName: xIdent.Name,
			FuncName:   v.Sel.Name,
		}, true
	case *ast.CallExpr:
		arg := v.Args[0]
		return getFuncType(arg)
	}
	panic("not exist type")
}

func (p *astParser) getAllGoFileInfoFromDepsByList(pkg *build.Package) error {
	ignoreInternal := pkg.Goroot && !p.ParseInternal
	if ignoreInternal { // ignored internal
		return nil
	}

	srcDir := pkg.Dir
	var err error
	for i := range pkg.GoFiles {
		err = p.parseFile(pkg.ImportPath, filepath.Join(srcDir, pkg.GoFiles[i]), nil)
		if err != nil {
			return err
		}
	}

	// parse .go source files that import "C"
	for i := range pkg.CgoFiles {
		err = p.parseFile(pkg.ImportPath, filepath.Join(srcDir, pkg.CgoFiles[i]), nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *astParser) parseFile(packageDir, path string, src interface{}) error {
	if strings.HasSuffix(strings.ToLower(path), "_test.go") || filepath.Ext(path) != ".go" {
		return nil
	}

	// positions are relative to FileSet
	astFile, err := parser.ParseFile(token.NewFileSet(), path, src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("ParseFile error:%+v", err)
	}

	err = p.packages.CollectAstFile(packageDir, path, astFile)
	if err != nil {
		return err
	}

	return nil
}

func listPackages(ctx context.Context, dir string, env []string, args ...string) (pkgs []*build.Package, finalErr error) {
	cmd := exec.CommandContext(ctx, "go", append([]string{"list", "-json", "-e"}, args...)...)
	cmd.Env = env
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	defer func() {
		if stderrBuf.Len() > 0 {
			finalErr = fmt.Errorf("%v\n%s", finalErr, stderrBuf.Bytes())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(stdout)
	for dec.More() {
		var pkg build.Package
		err = dec.Decode(&pkg)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, &pkg)
	}
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}
	return pkgs, nil
}

// ParseAPI parses general api info for given searchDir and mainAPIFile.
func (p *astParser) ParseAPI(searchDir string, mainAPIFile string, parseDepth int) error {
	return p.ParseAPIMultiSearchDir([]string{searchDir}, mainAPIFile, parseDepth)
}

// ParseAPIMultiSearchDir is like ParseAPI but for multiple search dirs.
func (p *astParser) ParseAPIMultiSearchDir(searchDirs []string, mainAPIFile string, parseDepth int) error {
	for _, searchDir := range searchDirs {
		elog.Info("Generate general API Info, search dir", elog.String("searchDir", searchDir))

		packageDir, err := getPkgName(searchDir)
		if err != nil {
			elog.Error("warning: failed to get package name", elog.String("searchDir", searchDir), elog.FieldErr(err))
		}

		err = p.getAllGoFileInfo(packageDir, searchDir)
		if err != nil {
			return err
		}
	}

	absMainAPIFilePath, err := filepath.Abs(filepath.Join(searchDirs[0], mainAPIFile))
	if err != nil {
		return err
	}

	// Use 'go list' command
	if p.ParseAllDependency {
		pkgs, err := listPackages(context.Background(), filepath.Dir(absMainAPIFilePath), nil, "-deps")
		if err != nil {
			return fmt.Errorf("pkg %s cannot find all dependencies, %s", filepath.Dir(absMainAPIFilePath), err)
		}
		for _, value := range pkgs {
			err := p.getAllGoFileInfoFromDepsByList(value)
			if err != nil {
				return err
			}
		}
	} else {
		if len(p.Dependences) > 0 {
			pkgs, err := listPackages(context.Background(), filepath.Dir(absMainAPIFilePath), nil, "-deps")
			if err != nil {
				return fmt.Errorf("pkg %s cannot find all dependencies, %s", filepath.Dir(absMainAPIFilePath), err)
			}
			for _, value := range pkgs {
				for _, prefix := range p.Dependences {
					//如果不在这个依赖里面直接过滤吊
					if !strings.HasPrefix(value.ImportPath, prefix) {
						continue
					}
					err := p.getAllGoFileInfoFromDepsByList(value)
					if err != nil {
						return err
					}
				}

			}
		}

	}
	//
	//err = p.ParseGeneralAPIInfo(absMainAPIFilePath)
	//if err != nil {
	//	return err
	//}

	p.parsedSchemas, err = p.packages.ParseTypes()
	if err != nil {
		return err
	}
	//fmt.Printf("p.parsedSchemas--------------->"+"%+v\n", p.parsedSchemas)
	//fmt.Printf("	p.parsedSchemas--------------->"+"%+v\n", len(p.packages.files))
	//for _, value := range p.packages.files {
	//	fmt.Printf("value--------------->"+"%+v\n", value)
	//}
	//err = rangeFiles(p.packages.files, p.ParseRouterAPIInfo)
	//if err != nil {
	//	return err
	//}

	//parser.renameRefSchemas()
	//
	//return parser.checkOperationIDUniqueness()
	return nil
}

// ParseRouterAPIInfo parses router api info for given astFile.
//func (p *astParser) ParseRouterAPIInfo(fileName string, astFile *ast.File) error {
//	for _, astDescription := range astFile.Decls {
//		astDeclaration, ok := astDescription.(*ast.FuncDecl)
//		if ok && astDeclaration.Doc != nil && astDeclaration.Doc.List != nil {
//			// for per 'function' comment, create a new 'Operation' object
//			operation := NewOperation(p, SetCodeExampleFilesDirectory(parser.codeExampleFilesDir))
//			for _, comment := range astDeclaration.Doc.List {
//				err := operation.ParseComment(comment.Text, astFile)
//				if err != nil {
//					return fmt.Errorf("ParseComment error in file %s :%+v", fileName, err)
//				}
//			}
//
//			err := processRouterOperation(parser, operation)
//			if err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}

// GetAllGoFileInfo gets all Go source files information for given searchDir.
func (p *astParser) getAllGoFileInfo(packageDir, searchDir string) error {
	return filepath.Walk(searchDir, func(path string, f os.FileInfo, _ error) error {
		err := p.Skip(path, f)
		if err != nil {
			return err
		}

		if f.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(searchDir, path)
		if err != nil {
			return err
		}
		// pkg/shop/shop.go

		return p.parseFile(filepath.ToSlash(filepath.Dir(filepath.Clean(filepath.Join(packageDir, relPath)))), path, nil)
	})
}

// Skip returns filepath.SkipDir error if match vendor and hidden folder.
func (p *astParser) Skip(path string, f os.FileInfo) error {
	return walkWith(p.excludes, p.ParseVendor)(path, f)
}

func walkWith(excludes map[string]struct{}, parseVendor bool) func(path string, fileInfo os.FileInfo) error {
	return func(path string, f os.FileInfo) error {
		if f.IsDir() {
			if !parseVendor && f.Name() == "vendor" || // ignore "vendor"
				f.Name() == "docs" || // exclude docs
				len(f.Name()) > 1 && f.Name()[0] == '.' { // exclude all hidden folder
				return filepath.SkipDir
			}

			if excludes != nil {
				if _, ok := excludes[path]; ok {
					return filepath.SkipDir
				}
			}
		}

		return nil
	}
}
func getPkgName(searchDir string) (string, error) {
	cmd := exec.Command("go", "list", "-f={{.PackagePath}}")
	cmd.Dir = searchDir

	var stdout, stderr strings.Builder

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("execute go list command, %s, stdout:%s, stderr:%s", err, stdout.String(), stderr.String())
	}

	outStr, _ := stdout.String(), stderr.String()

	if outStr[0] == '_' { // will shown like _/{GOPATH}/src/{YOUR_PACKAGE} when NOT enable GO MODULE.
		outStr = strings.TrimPrefix(outStr, "_"+build.Default.GOPATH+"/src/")
	}

	f := strings.Split(outStr, "\n")

	outStr = f[0]

	return outStr, nil
}

// ParseGeneralAPIInfo parses general api info for given mainAPIFile path.
func (p *astParser) ParseGeneralAPIInfo(mainAPIFile string) error {
	fileTree, err := parser.ParseFile(token.NewFileSet(), mainAPIFile, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("cannot parse source files %s: %s", mainAPIFile, err)
	}

	for _, comment := range fileTree.Comments {
		comments := strings.Split(comment.Text(), "\n")
		if !isGeneralAPIComment(comments) {
			continue
		}

		err = parseGeneralAPIInfo(p, comments)
		if err != nil {
			return err
		}
	}

	return nil
}

func isGeneralAPIComment(comments []string) bool {
	//for _, commentLine := range comments {
	//	attribute := strings.ToLower(strings.Split(commentLine, " ")[0])
	//	switch attribute {
	//	// The @summary, @router, @success, @failure annotation belongs to Operation
	//	case summaryAttr, routerAttr, successAttr, failureAttr, responseAttr:
	//		return false
	//	}
	//}

	return true
}
func parseGeneralAPIInfo(parser *astParser, comments []string) error {
	//previousAttribute := ""

	// parsing classic meta data model
	//for line := 0; line < len(comments); line++ {
	//	commentLine := comments[line]
	//	attribute := strings.Split(commentLine, " ")[0]
	//	value := strings.TrimSpace(commentLine[len(attribute):])
	//
	//	multilineBlock := false
	//	if previousAttribute == attribute {
	//		multilineBlock = true
	//	}
	//
	//	switch attr := strings.ToLower(attribute); attr {
	//	case versionAttr, titleAttr, tosAttr, licNameAttr, licURLAttr, conNameAttr, conURLAttr, conEmailAttr:
	//		setSwaggerInfo(parser.swagger, attr, value)
	//	case descriptionAttr:
	//		if multilineBlock {
	//			parser.swagger.Info.Description += "\n" + value
	//
	//			continue
	//		}
	//
	//		setSwaggerInfo(parser.swagger, attr, value)
	//	case descriptionMarkdownAttr:
	//		commentInfo, err := getMarkdownForTag("api", parser.markdownFileDir)
	//		if err != nil {
	//			return err
	//		}
	//
	//		setSwaggerInfo(parser.swagger, descriptionAttr, string(commentInfo))
	//
	//	case "@host":
	//		parser.swagger.Host = value
	//	case "@basepath":
	//		parser.swagger.BasePath = value
	//
	//	case acceptAttr:
	//		err := parser.ParseAcceptComment(value)
	//		if err != nil {
	//			return err
	//		}
	//	case produceAttr:
	//		err := parser.ParseProduceComment(value)
	//		if err != nil {
	//			return err
	//		}
	//	case "@schemes":
	//		parser.swagger.Schemes = strings.Split(value, " ")
	//	case "@tag.name":
	//		parser.swagger.Tags = append(parser.swagger.Tags, spec.Tag{
	//			TagProps: spec.TagProps{
	//				Name: value,
	//			},
	//		})
	//	case "@tag.description":
	//		tag := parser.swagger.Tags[len(parser.swagger.Tags)-1]
	//		tag.TagProps.Description = value
	//		replaceLastTag(parser.swagger.Tags, tag)
	//	case "@tag.description.markdown":
	//		tag := parser.swagger.Tags[len(parser.swagger.Tags)-1]
	//
	//		commentInfo, err := getMarkdownForTag(tag.TagProps.Name, parser.markdownFileDir)
	//		if err != nil {
	//			return err
	//		}
	//
	//		tag.TagProps.Description = string(commentInfo)
	//		replaceLastTag(parser.swagger.Tags, tag)
	//	case "@tag.docs.url":
	//		tag := parser.swagger.Tags[len(parser.swagger.Tags)-1]
	//		tag.TagProps.ExternalDocs = &spec.ExternalDocumentation{
	//			URL:         value,
	//			Description: "",
	//		}
	//
	//		replaceLastTag(parser.swagger.Tags, tag)
	//	case "@tag.docs.description":
	//		tag := parser.swagger.Tags[len(parser.swagger.Tags)-1]
	//		if tag.TagProps.ExternalDocs == nil {
	//			return fmt.Errorf("%s needs to come after a @tags.docs.url", attribute)
	//		}
	//
	//		tag.TagProps.ExternalDocs.Description = value
	//		replaceLastTag(parser.swagger.Tags, tag)
	//
	//	case secBasicAttr, secAPIKeyAttr, secApplicationAttr, secImplicitAttr, secPasswordAttr, secAccessCodeAttr:
	//		scheme, err := parseSecAttributes(attribute, comments, &line)
	//		if err != nil {
	//			return err
	//		}
	//
	//		parser.swagger.SecurityDefinitions[value] = scheme
	//
	//	case "@query.collection.format":
	//		parser.collectionFormatInQuery = value
	//	default:
	//		prefixExtension := "@x-"
	//		// Prefix extension + 1 char + 1 space  + 1 char
	//		if len(attribute) > 5 && attribute[:len(prefixExtension)] == prefixExtension {
	//			extExistsInSecurityDef := false
	//			// for each security definition
	//			for _, v := range parser.swagger.SecurityDefinitions {
	//				// check if extension exists
	//				_, extExistsInSecurityDef = v.VendorExtensible.Extensions.GetString(attribute[1:])
	//				// if it exists in at least one, then we stop iterating
	//				if extExistsInSecurityDef {
	//					break
	//				}
	//			}
	//
	//			// if it is present on security def, don't add it again
	//			if extExistsInSecurityDef {
	//				break
	//			}
	//
	//			var valueJSON interface{}
	//
	//			split := strings.SplitAfter(commentLine, attribute+" ")
	//			if len(split) < 2 {
	//				return fmt.Errorf("annotation %s need a value", attribute)
	//			}
	//
	//			extensionName := "x-" + strings.SplitAfter(attribute, prefixExtension)[1]
	//
	//			err := json.Unmarshal([]byte(split[1]), &valueJSON)
	//			if err != nil {
	//				return fmt.Errorf("annotation %s need a valid json value", attribute)
	//			}
	//
	//			if strings.Contains(extensionName, "logo") {
	//				parser.swagger.Info.Extensions.Add(extensionName, valueJSON)
	//			} else {
	//				if parser.swagger.Extensions == nil {
	//					parser.swagger.Extensions = make(map[string]interface{})
	//				}
	//
	//				parser.swagger.Extensions[attribute[1:]] = valueJSON
	//			}
	//		}
	//	}
	//
	//	previousAttribute = attribute
	//}

	return nil
}

// GetSchemaTypePath get Path of schema type.
func (p *astParser) GetSchemaTypePath(schema *spec.Schema, depth int) []string {
	if schema == nil || depth == 0 {
		return nil
	}
	name := schema.Ref.String()
	if name != "" {
		if pos := strings.LastIndexByte(name, '/'); pos >= 0 {
			name = name[pos+1:]
			if schema, ok := p.swagger.Definitions[name]; ok {
				return p.GetSchemaTypePath(&schema, depth)
			}
		}

		return nil
	}
	if len(schema.Type) > 0 {
		switch schema.Type[0] {
		case ARRAY:
			depth--
			s := []string{schema.Type[0]}

			return append(s, p.GetSchemaTypePath(schema.Items.Schema, depth)...)
		case OBJECT:
			if schema.AdditionalProperties != nil && schema.AdditionalProperties.Schema != nil {
				// for map
				depth--
				s := []string{schema.Type[0]}

				return append(s, p.GetSchemaTypePath(schema.AdditionalProperties.Schema, depth)...)
			}
		}

		return []string{schema.Type[0]}
	}

	return []string{ANY}
}
