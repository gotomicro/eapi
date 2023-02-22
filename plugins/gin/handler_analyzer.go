package gin

import (
	"fmt"
	"go/ast"
	"log"
	"net/http"
	"strconv"
	"strings"

	analyzer "github.com/gotomicro/eapi"
	"github.com/gotomicro/eapi/plugins/common"
	"github.com/gotomicro/eapi/spec"
	"github.com/gotomicro/eapi/utils"
	"github.com/iancoleman/strcase"
	"github.com/robertkrimen/otto"
	"github.com/samber/lo"
)

const ginContextIdentName = "*github.com/gin-gonic/gin.Context"

var (
	interestedGinContextMethods = []string{
		"Bind",
		"BindJSON",
		"BindXML",
		"BindYAML",
		"BindTOML",
		"BindUri",
		"ShouldBind",
		"ShouldBindJSON",
		"ShouldBindXML",
		"ShouldBindYAML",
		"ShouldBindTOML",
		"ShouldBindUri",
		"ShouldBindHeader",
		"JSON",
		"Query",
		"Param",
		"PostForm",
		"PostFormArray",
		"GetPostForm",
		"GetPostFormArray",
		"XML",
		"Redirect",
		"FormFile",
		"DefaultQuery",
		"DefaultPostForm",
	}
)

type handlerAnalyzer struct {
	ctx  *analyzer.Context
	api  *analyzer.API
	spec *analyzer.APISpec
	decl *ast.FuncDecl

	c *common.Config
}

func newHandlerParser(ctx *analyzer.Context, api *analyzer.API, decl *ast.FuncDecl) *handlerAnalyzer {
	return &handlerAnalyzer{ctx: ctx, api: api, spec: api.Spec, decl: decl}
}

func (p *handlerAnalyzer) WithConfig(c *common.Config) *handlerAnalyzer {
	p.c = c
	return p
}

func (p *handlerAnalyzer) Parse() {
	ast.Inspect(p.decl, func(node ast.Node) bool {
		matched := p.matchCustomResponseRule(node)
		if matched {
			return true
		}
		matched = p.matchCustomRequestRule(node)
		if matched {
			return true
		}

		p.ctx.MatchCall(node,
			analyzer.NewCallRule().WithRule(ginContextIdentName, interestedGinContextMethods...),
			func(call *ast.CallExpr, typeName, fnName string) {
				switch fnName {
				case "Bind", "ShouldBind":
					p.parseBinding(call)
				case "BindJSON", "ShouldBindJSON":
					p.parseBindWithContentType(call, analyzer.MimeTypeJson)
				case "BindXML", "ShouldBindXML":
					p.parseBindWithContentType(call, analyzer.MimeApplicationXml)
				case "BindYAML", "ShouldBindYAML":
					p.parseBindWithContentType(call, "application/yaml")
				case "BindTOML", "ShouldBindTOML":
					p.parseBindWithContentType(call, "application/toml")
				case "BindUri", "ShouldBindUri":
					// TODO
				case "BindHeader", "ShouldBindHeader":
					// TODO
				case "JSON":
					p.parseResBody(call, analyzer.MimeTypeJson)
				case "XML":
					p.parseResBody(call, analyzer.MimeApplicationXml)
				case "Query": // query parameter
					p.parsePrimitiveParam(call, "query")
				case "Param": // path parameter
					p.parsePrimitiveParam(call, "path")
				case "PostForm", "GetPostForm":
					p.parseFormData(call, "string")
				case "FormFile":
					p.parseFormData(call, "file")
				case "PostFormArray", "GetPostFormArray":
					p.parseFormData(call, spec.TypeArray, func(s *spec.Schema) {
						s.Items = spec.NewSchemaRef("", spec.NewStringSchema())
					})
				case "Redirect":
					p.parseRedirectRes(call)
				case "DefaultQuery":
					p.parsePrimitiveParamWithDefault(call, "query")
				case "DefaultPostForm":
					p.parseFormData(call, "string")
					// TODO: supporting more methods (FileForm(), HTML(), Data(), etc...)
				}
			},
		)
		return true
	})
}

func (p *handlerAnalyzer) paramNameParser(fieldName string, tags map[string]string) (name, in string) {
	name, ok := tags["form"]
	if ok {
		return name, "query"
	}
	return fieldName, "query"
}

func (p *handlerAnalyzer) parseBinding(call *ast.CallExpr) {
	if len(call.Args) != 1 {
		return
	}
	arg0 := call.Args[0]

	switch p.api.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodDelete:
		params := analyzer.NewParamParser(p.ctx, p.paramNameParser).Parse(arg0)
		for _, param := range params {
			p.spec.AddParameter(param)
		}
	default:
		contentType := p.getRequestContentType("")
		p.parseBindWithContentType(call, contentType)
	}
}

func (p *handlerAnalyzer) parseBindWithContentType(call *ast.CallExpr, contentType string) {
	if len(call.Args) != 1 {
		return
	}
	arg0 := call.Args[0]

	schema := p.ctx.GetSchemaByExpr(arg0, contentType)
	if schema == nil {
		return
	}
	commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
	if commentGroup != nil {
		comment := p.ctx.ParseComment(commentGroup)
		schema.Description = comment.Text()
	}
	reqBody := spec.NewRequestBody().WithSchemaRef(schema, []string{contentType})
	p.spec.RequestBody = &spec.RequestBodyRef{Value: reqBody}
}

func (p *handlerAnalyzer) parseResBody(call *ast.CallExpr, contentType string) {
	if len(call.Args) != 2 {
		return
	}

	res := spec.NewResponse()
	commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
	if commentGroup != nil {
		comment := p.ctx.ParseComment(commentGroup)
		res.Description = comment.TextPointer()
	}

	schema := p.ctx.GetSchemaByExpr(call.Args[1], contentType)
	res.Content = spec.NewContentWithSchemaRef(schema, []string{contentType})
	statusCode := p.ctx.ParseStatusCode(call.Args[0])
	p.spec.AddResponse(statusCode, res)
}

func (p *handlerAnalyzer) parseRedirectRes(call *ast.CallExpr) {
	if len(call.Args) == 0 {
		return
	}

	res := spec.NewResponse()
	commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
	if commentGroup != nil {
		comment := p.ctx.ParseComment(commentGroup)
		if comment != nil {
			desc := comment.Text()
			res.Description = &desc
		}
	}
	statusCode := p.ctx.ParseStatusCode(call.Args[0])
	p.spec.AddResponse(statusCode, res)
}

func (p *handlerAnalyzer) parsePrimitiveParam(call *ast.CallExpr, in string) {
	param := p.primitiveParam(call, in)
	p.spec.AddParameter(param)
}

func (p *handlerAnalyzer) parsePrimitiveParamWithDefault(call *ast.CallExpr, in string) {
	param := p.primitiveParamWithDefault(call, in)
	p.spec.AddParameter(param)
}

func (p *handlerAnalyzer) parseFormData(call *ast.CallExpr, fieldType string, options ...func(s *spec.Schema)) {
	if len(call.Args) <= 0 {
		return
	}
	arg0 := call.Args[0]
	arg0Lit, ok := arg0.(*ast.BasicLit)
	if !ok {
		return
	}

	name := strings.Trim(arg0Lit.Value, "\"")
	paramSchema := spec.NewSchema()
	paramSchema.Title = name
	paramSchema.Type = fieldType
	for _, option := range options {
		option(paramSchema)
	}

	requestBody := p.spec.RequestBody
	if requestBody == nil {
		requestBody = &spec.RequestBodyRef{Value: spec.NewRequestBody().WithContent(spec.NewContent())}
		p.spec.RequestBody = requestBody
	}
	mediaType := requestBody.Value.GetMediaType(analyzer.MimeTypeFormData)
	if mediaType == nil {
		mediaType = spec.NewMediaType()
		requestBody.Value.Content[analyzer.MimeTypeFormData] = mediaType
	}

	comment := p.ctx.ParseComment(p.ctx.GetHeadingCommentOf(call.Pos()))
	paramSchema.Description = comment.Text()

	var schemaRef = mediaType.Schema
	var schema *spec.SchemaRef
	if schemaRef != nil {
		schema = spec.Unref(p.ctx.Doc(), schemaRef)
		schema.Value.WithProperty(name, paramSchema)
	} else {
		schema = spec.NewObjectSchema().NewRef()
		title := strcase.ToCamel(p.spec.OperationID) + "Request"
		schema.Value.Title = title
		schema.Value.WithProperty(name, paramSchema)
		p.ctx.Doc().Components.Schemas[title] = schema
		schemaRef = spec.RefComponentSchemas(title)
		mediaType.Schema = schemaRef
	}
	if comment.Required() {
		schema.Value.Required = append(schema.Value.Required, name)
	}
}

func (p *handlerAnalyzer) primitiveParamWithDefault(call *ast.CallExpr, in string) *spec.Parameter {
	if len(call.Args) < 2 {
		return nil
	}
	param := p.primitiveParam(call, in)
	if param == nil {
		return nil
	}

	arg1Lit, ok := call.Args[1].(*ast.BasicLit)
	if !ok {
		return nil
	}
	param.Schema.Value.Default, _ = strconv.Unquote(arg1Lit.Value)

	return param
}

func (p *handlerAnalyzer) primitiveParam(call *ast.CallExpr, in string) *spec.Parameter {
	if len(call.Args) <= 0 {
		return nil
	}
	arg0 := call.Args[0]
	arg0Lit, ok := arg0.(*ast.BasicLit)
	if !ok {
		return nil
	}
	name := strings.Trim(arg0Lit.Value, "\"")
	paramSchema := spec.NewSchema()
	paramSchema.Title = name
	paramSchema.Type = "string"

	comment := p.ctx.ParseComment(p.ctx.GetHeadingCommentOf(call.Pos()))

	var res *spec.Parameter
	switch in {
	case "path":
		res = spec.NewPathParameter(name).WithSchema(paramSchema)
	case "query":
		res = spec.NewQueryParameter(name).WithSchema(paramSchema)
		res.Required = comment.Required()
	default:
		return nil
	}

	res.Description = comment.Text()
	return res
}

func (p *handlerAnalyzer) matchCustomResponseRule(node ast.Node) (matched bool) {
	if p.c == nil || len(p.c.Response) == 0 {
		return false
	}

	for _, rule := range p.c.Response {
		p.ctx.MatchCall(
			node,
			analyzer.NewCallRule().WithRule(rule.Type, rule.Method),
			func(call *ast.CallExpr, typeName, fnName string) {
				matched = true

				var contentType = rule.Return.ContentType
				//if !lo.Contains(p.spec.Produces, contentType) {
				//	p.spec.Produces = append(p.spec.Produces, contentType)
				//}

				res := spec.NewResponse()
				comment := p.ctx.ParseComment(p.ctx.GetHeadingCommentOf(call.Pos()))
				res.Description = comment.TextPointer()
				schema := p.parseDataType(call, rule.Return.Data, contentType)
				res.WithContent(spec.NewContentWithSchemaRef(schema, []string{contentType}))
				statusCode := p.parseStatusCodeInCall(call, rule.Return.Status)
				p.spec.AddResponse(statusCode, res)
			},
		)
	}

	return
}

func (p *handlerAnalyzer) matchCustomRequestRule(node ast.Node) (matched bool) {
	if p.c == nil || len(p.c.Request) == 0 {
		return false
	}

	for _, rule := range p.c.Request {
		p.ctx.MatchCall(
			node,
			analyzer.NewCallRule().WithRule(rule.Type, rule.Method),
			func(call *ast.CallExpr, typeName, fnName string) {
				matched = true

				switch p.api.Method {
				case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodDelete:
					params := p.parseParamsInCall(call, rule.Return.Data, analyzer.MimeTypeFormData)
					for _, param := range params {
						p.spec.AddParameter(param)
					}

				default:
					contentType := p.getRequestContentType(rule.Return.ContentType)
					schema := p.parseDataType(call, rule.Return.Data, contentType)
					if schema == nil {
						return
					}
					reqBody := spec.NewRequestBody()
					reqBody.Required = true
					commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
					if commentGroup != nil {
						comment := p.ctx.ParseComment(commentGroup)
						reqBody.Description = comment.Text()
					}
					reqBody.WithSchemaRef(schema, []string{contentType})
					p.spec.RequestBody = &spec.RequestBodyRef{Value: reqBody}
				}

			},
		)
	}

	return
}

func (p *handlerAnalyzer) getFormDataIn() string {
	var in string
	switch p.api.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodDelete:
		in = "query"
	default:
		in = "form"
	}
	return in
}

func (p *handlerAnalyzer) parseStatusCodeInCall(call *ast.CallExpr, statusCode string) (code int) {
	if statusCode == "" {
		return 200 // default to 200
	}

	output := p.evaluate(call, statusCode)
	switch value := output.(type) {
	case int64:
		code = int(value)
	case int:
		code = value
	case ast.Expr:
		code = p.ctx.ParseStatusCode(value)
	}

	return
}

func (p *handlerAnalyzer) parseDataType(call *ast.CallExpr, dataType *common.DataSchema, contentType string) (schema *spec.SchemaRef) {
	switch dataType.Type {
	case common.DataTypeString:
		return p.basicSchemaType("string")
	case common.DataTypeNumber:
		return p.basicSchemaType("number")
	case common.DataTypeInteger:
		return p.basicSchemaType("integer")
	case common.DataTypeBoolean:
		return p.basicSchemaType("boolean")
	case common.DataTypeFile:
		return p.basicSchemaType("file")
	case common.DataTypeArray:
		return spec.NewArraySchema(p.parseDataType(call, dataType.Item, contentType)).NewRef()
	case common.DataTypeObject:
		schema := spec.NewObjectSchema()
		properties := make(spec.Schemas)
		utils.RangeMapInOrder(
			dataType.Properties,
			func(a, b string) bool { return a < b },
			func(name string, dataSchema *common.DataSchema) {
				if !dataSchema.Optional {
					schema.Required = append(schema.Required, name)
				}
				s := p.parseDataType(call, dataSchema, contentType)
				if s != nil {
					properties[name] = s
				}
			},
		)
		schema.Properties = properties
		return spec.NewSchemaRef("", schema)
	default: // fallback to js expression
		output := p.evaluate(call, string(dataType.Type))
		if output == nil {
			return nil
		}
		expr, ok := output.(ast.Expr)
		if !ok {
			fmt.Printf("invalid data type '%s' in configuration file\n", dataType.Type)
			return nil
		}
		return p.ctx.GetSchemaByExpr(expr, contentType)
	}
}

func (p *handlerAnalyzer) basicSchemaType(t string) *spec.SchemaRef {
	return &spec.SchemaRef{
		Value: &spec.Schema{
			Type: t,
		},
	}
}

func (p *handlerAnalyzer) parseParamsInCall(call *ast.CallExpr, dataType *common.DataSchema, contentType string) (params []*spec.Parameter) {
	switch dataType.Type {
	case common.DataTypeString, common.DataTypeNumber, common.DataTypeInteger, common.DataTypeBoolean, common.DataTypeFile, common.DataTypeArray:
		param := &spec.Parameter{}
		schema := spec.NewSchema()
		schema.Type = string(dataType.Type)
		schema.Format = dataType.Format
		param.Schema = spec.NewSchemaRef("", schema)
		return append(params, param)

	case common.DataTypeObject: // unsupported in form data
		fmt.Printf("object is unsupported in form data\n")
		return

	default:
		output := p.evaluate(call, string(dataType.Type))
		expr, ok := output.(ast.Expr)
		if !ok {
			fmt.Printf("invalid data type '%s' in configuration file\n", dataType.Type)
			return nil
		}
		return analyzer.NewParamParser(p.ctx, p.paramNameParser).Parse(expr)
	}
}

func (p *handlerAnalyzer) evaluate(call *ast.CallExpr, code string) interface{} {
	env := otto.New()
	_ = env.Set("args", call.Args)
	output, err := env.Run(code)
	if err != nil {
		log.Fatalln("evaluate failed", err)
	}

	value, err := output.Export()
	if err != nil {
		log.Fatalln("evaluate failed", err)
	}

	return value
}

// 获取一个尽可能正确的 request payload contentType
func (p *handlerAnalyzer) getRequestContentType(contentType string) string {
	if contentType != "" {
		if !lo.Contains(p.spec.Consumes, contentType) {
			p.spec.Consumes = append(p.spec.Consumes, contentType)
		}
		return contentType
	}
	if len(p.spec.Consumes) != 0 {
		return p.spec.Consumes[0]
	}

	// fallback
	switch p.api.Method {
	case http.MethodGet, http.MethodHead:
		return analyzer.MimeTypeFormData
	default:
		return analyzer.MimeTypeJson
	}
}
