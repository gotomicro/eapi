package gin

import (
	"fmt"
	"go/ast"
	"log"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	analyzer "github.com/gotomicro/ego-gen-api"
	"github.com/gotomicro/ego-gen-api/spec"
	"github.com/robertkrimen/otto"
	"github.com/samber/lo"
)

const ginContextIdentName = "*github.com/gin-gonic/gin.Context"

var (
	interestedGinContextMethods = []string{"Bind", "JSON", "Query", "Param", "GetPostForm", "PostFormArray", "XML", "Redirect", "FormFile"}
)

type HandlerParser struct {
	ctx  *analyzer.Context
	api  *analyzer.API
	spec *analyzer.APISpec
	decl *ast.FuncDecl

	c *Config
}

func NewHandlerParser(ctx *analyzer.Context, api *analyzer.API, decl *ast.FuncDecl) *HandlerParser {
	return &HandlerParser{ctx: ctx, api: api, spec: api.Spec, decl: decl}
}

func (p *HandlerParser) WithConfig(c *Config) *HandlerParser {
	p.c = c
	return p
}

func (p *HandlerParser) Parse() {
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
				case "Bind":
					p.parseBinding(call)
				case "JSON":
					p.parseResBody(call, "application/json")
				case "XML":
					p.parseResBody(call, "application/xml")
				case "Query": // query parameter
					p.parsePrimitiveParam(call, "query")
				case "Param": // path parameter
					p.parsePrimitiveParam(call, "path")
				case "GetPostForm":
					p.parseFormData(call, "string")
				case "FormFile":
					p.parseFormData(call, "file")
				case "PostFormArray":
					p.parsePrimitiveParamArray(call, "formData")
				case "Redirect":
					p.parseRedirectRes(call)
					// TODO: supporting more methods (FileForm(), HTML(), Data(), etc...)
				}
			},
		)
		return true
	})
}

func (p *HandlerParser) parseBinding(call *ast.CallExpr) {
	if len(call.Args) != 1 {
		return
	}
	arg0 := call.Args[0]

	switch p.api.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodDelete:
		params := analyzer.NewParamParser(p.ctx, analyzer.MimeTypeFormData).Parse(arg0)
		for _, param := range params {
			param.In = "query"
			p.spec.AddParameter(param)
		}
	default:
		contentType := p.getRequestContentType("")
		schema := p.ctx.GetSchemaByExpr(arg0, contentType)
		if schema == nil {
			return
		}
		commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
		if commentGroup != nil {
			comment := analyzer.ParseComment(commentGroup)
			if comment != nil {
				schema.Description = comment.Text
			}
		}
		reqBody := openapi3.NewRequestBody().WithSchemaRef(schema, []string{contentType})
		p.spec.RequestBody = &openapi3.RequestBodyRef{Value: reqBody}
	}
}

func (p *HandlerParser) parseResBody(call *ast.CallExpr, contentType string) {
	if len(call.Args) != 2 {
		return
	}

	res := openapi3.NewResponse()
	commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
	if commentGroup != nil {
		comment := analyzer.ParseComment(commentGroup)
		if comment != nil {
			res.Description = &comment.Text
		}
	}

	schema := p.ctx.GetSchemaByExpr(call.Args[1], contentType)
	res.Content = openapi3.NewContentWithSchemaRef(schema, []string{contentType})
	statusCode := p.ctx.ParseStatusCode(call.Args[0])
	p.spec.AddResponse(statusCode, res)
}

func (p *HandlerParser) parseRedirectRes(call *ast.CallExpr) {
	if len(call.Args) == 0 {
		return
	}

	res := openapi3.NewResponse()
	commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
	if commentGroup != nil {
		comment := analyzer.ParseComment(commentGroup)
		if comment != nil {
			res.Description = &comment.Text
		}
	}
	statusCode := p.ctx.ParseStatusCode(call.Args[0])
	p.spec.AddResponse(statusCode, res)
}

func (p *HandlerParser) parsePrimitiveParam(call *ast.CallExpr, in string) {
	param := p.primitiveParam(call, in)
	p.spec.AddParameter(param)
}

func (p *HandlerParser) parseFormData(call *ast.CallExpr, fieldType string) {
	if len(call.Args) <= 0 {
		return
	}
	arg0 := call.Args[0]
	arg0Lit, ok := arg0.(*ast.BasicLit)
	if !ok {
		return
	}

	name := strings.Trim(arg0Lit.Value, "\"")
	paramSchema := openapi3.NewSchema()
	paramSchema.Title = name
	paramSchema.Type = fieldType

	mediaType := openapi3.NewMediaType()
	requestBody := p.spec.RequestBody
	if requestBody == nil {
		requestBody = &openapi3.RequestBodyRef{Value: openapi3.NewRequestBody().WithContent(openapi3.NewContent())}
		p.spec.RequestBody = requestBody
	}
	if requestBody.Value.GetMediaType(analyzer.MimeTypeFormData) != nil {
		mediaType = p.spec.RequestBody.Value.GetMediaType(analyzer.MimeTypeFormData)
	}

	var schema *openapi3.SchemaRef
	if mediaType.Schema != nil {
		schema = spec.Unref(p.ctx.Doc(), mediaType.Schema)
		schema.Value.WithProperty(name, paramSchema)
	} else {
		schema = openapi3.NewSchemaRef("", openapi3.NewObjectSchema())
		schema.Value.WithProperty(name, paramSchema)
		mediaType.Schema = schema
		p.spec.RequestBody.Value.Content[analyzer.MimeTypeFormData] = mediaType
	}

	comment := analyzer.ParseComment(p.ctx.GetHeadingCommentOf(call.Lparen))
	if comment.Required() {
		schema.Value.Required = append(schema.Value.Required, name)
	}
}

func (p *HandlerParser) primitiveParam(call *ast.CallExpr, in string) *openapi3.Parameter {
	if len(call.Args) <= 0 {
		return nil
	}
	arg0 := call.Args[0]
	arg0Lit, ok := arg0.(*ast.BasicLit)
	if !ok {
		return nil
	}
	name := strings.Trim(arg0Lit.Value, "\"")
	paramSchema := openapi3.NewSchema()
	paramSchema.Title = name

	res := openapi3.NewPathParameter(name).WithSchema(paramSchema)
	commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
	if commentGroup != nil {
		comment := analyzer.ParseComment(commentGroup)
		if comment != nil {
			res.Description = comment.Text
		}
	}

	return res
}

func (p *HandlerParser) parsePrimitiveParamArray(call *ast.CallExpr, in string) {
	param := p.primitiveParam(call, in)
	param.Schema.Value.Type = "array"
	param.Schema.Value.Items = openapi3.NewSchemaRef("", openapi3.NewStringSchema())
	p.spec.AddParameter(param)
}

func (p *HandlerParser) matchCustomResponseRule(node ast.Node) (matched bool) {
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

				res := openapi3.NewResponse()
				commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
				if commentGroup != nil {
					comment := analyzer.ParseComment(commentGroup)
					if comment != nil {
						res.Description = &comment.Text
					}
				}
				schema := p.parseDataType(call, rule.Return.Data, contentType)
				res.WithContent(openapi3.NewContentWithSchemaRef(schema, []string{contentType}))
				statusCode := p.parseStatusCodeInCall(call, rule.Return.Status)
				p.spec.AddResponse(statusCode, res)
			},
		)
	}

	return
}

func (p *HandlerParser) matchCustomRequestRule(node ast.Node) (matched bool) {
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
						param.In = "query"
						p.spec.AddParameter(param)
					}

				default:
					contentType := p.getRequestContentType(rule.Return.ContentType)
					schema := p.parseDataType(call, rule.Return.Data, contentType)
					if schema == nil {
						return
					}
					reqBody := openapi3.NewRequestBody()
					reqBody.Required = true
					commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
					if commentGroup != nil {
						comment := analyzer.ParseComment(commentGroup)
						if comment != nil {
							reqBody.Description = comment.Text
						}
					}
					reqBody.WithSchemaRef(schema, []string{contentType})
					p.spec.RequestBody = &openapi3.RequestBodyRef{Value: reqBody}
				}

			},
		)
	}

	return
}

func (p *HandlerParser) getFormDataIn() string {
	var in string
	switch p.api.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodDelete:
		in = "query"
	default:
		in = "form"
	}
	return in
}

func (p *HandlerParser) parseStatusCodeInCall(call *ast.CallExpr, statusCode string) (code int) {
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

func (p *HandlerParser) parseDataType(call *ast.CallExpr, dataType *DataSchema, contentType string) (schema *openapi3.SchemaRef) {
	switch dataType.Type {
	case DataTypeString:
		return p.basicSchemaType("string")
	case DataTypeNumber:
		return p.basicSchemaType("number")
	case DataTypeInteger:
		return p.basicSchemaType("integer")
	case DataTypeBoolean:
		return p.basicSchemaType("boolean")
	case DataTypeFile:
		return p.basicSchemaType("file")
	case DataTypeArray:
		schema := openapi3.NewArraySchema()
		schema.Items = p.parseDataType(call, dataType.Item, contentType)
		return openapi3.NewSchemaRef("", schema)
	case DataTypeObject:
		schema := openapi3.NewObjectSchema()
		properties := make(openapi3.Schemas)
		for name, dataSchema := range dataType.Properties {
			s := p.parseDataType(call, dataSchema, contentType)
			if s != nil {
				properties[name] = s
			}
		}
		schema.Properties = properties
		return openapi3.NewSchemaRef("", schema)
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

func (p *HandlerParser) basicSchemaType(t string) *openapi3.SchemaRef {
	return &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: t,
		},
	}
}

func (p *HandlerParser) parseParamsInCall(call *ast.CallExpr, dataType *DataSchema, contentType string) (params []*openapi3.Parameter) {
	switch dataType.Type {
	case DataTypeString, DataTypeNumber, DataTypeInteger, DataTypeBoolean, DataTypeFile, DataTypeArray:
		param := &openapi3.Parameter{}
		schema := openapi3.NewSchema()
		schema.Type = string(dataType.Type)
		schema.Format = dataType.Format
		param.Schema = openapi3.NewSchemaRef("", schema)
		return append(params, param)

	case DataTypeObject: // unsupported in form data
		fmt.Printf("object is unsupported in form data\n")
		return

	default:
		output := p.evaluate(call, string(dataType.Type))
		expr, ok := output.(ast.Expr)
		if !ok {
			fmt.Printf("invalid data type '%s' in configuration file\n", dataType.Type)
			return nil
		}
		return analyzer.NewParamParser(p.ctx, contentType).Parse(expr)
	}
}

func (p *HandlerParser) evaluate(call *ast.CallExpr, code string) interface{} {
	env := otto.New()
	env.Set("args", call.Args)
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
func (p *HandlerParser) getRequestContentType(contentType string) string {
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
