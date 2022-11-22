package gin

import (
	"fmt"
	"go/ast"
	"log"
	"net/http"
	"strings"

	"github.com/go-openapi/spec"
	analyzer "github.com/gotomicro/ego-gen-api"
	"github.com/robertkrimen/otto"
	"github.com/samber/lo"
)

const ginContextIdentName = "*github.com/gin-gonic/gin.Context"

var (
	interestedGinContextMethods = []string{"Bind", "JSON", "Query", "Param", "GetPostForm", "PostFormArray", "XML", "Redirect"}
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
					p.parsePrimitiveParam(call, "formData")
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

	contentType := p.getRequestContentType("")
	switch contentType {
	case analyzer.MimeTypeFormData, analyzer.MimeTypeFormUrlencoded:
		in := p.getFormDataIn()
		params := analyzer.NewParamParser(p.ctx, contentType).Parse(arg0)
		for _, param := range params {
			param.In = in
			p.spec.AddParam(param)
		}

	default:
		schema := p.ctx.GetSchemaByExpr(arg0, contentType)
		if schema == nil {
			return
		}
		param := spec.BodyParam("payload", schema)

		commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
		if commentGroup != nil {
			comment := analyzer.ParseComment(commentGroup)
			if comment != nil {
				param.Description = comment.Text
			}
		}

		p.spec.AddParam(param)
	}
}

func (p *HandlerParser) parseResBody(call *ast.CallExpr, contentType string) {
	if len(call.Args) != 2 {
		return
	}
	if !lo.Contains(p.spec.Produces, contentType) {
		p.spec.Produces = append(p.spec.Produces, contentType)
	}

	res := spec.NewResponse()
	res.Schema = p.ctx.GetSchemaByExpr(call.Args[1], contentType)
	commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
	if commentGroup != nil {
		comment := analyzer.ParseComment(commentGroup)
		if comment != nil {
			res.Description = comment.Text
		}
	}

	statusCode := p.ctx.ParseStatusCode(call.Args[0])
	p.spec.RespondsWith(statusCode, res)
}

func (p *HandlerParser) parseRedirectRes(call *ast.CallExpr) {
	if len(call.Args) == 0 {
		return
	}

	res := spec.NewResponse()
	commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
	if commentGroup != nil {
		comment := analyzer.ParseComment(commentGroup)
		if comment != nil {
			res.Description = comment.Text
		}
	}
	statusCode := p.ctx.ParseStatusCode(call.Args[0])
	p.spec.RespondsWith(statusCode, res)
}

func (p *HandlerParser) parsePrimitiveParam(call *ast.CallExpr, in string) {
	param := p.primitiveParam(call, in)
	p.spec.AddParam(param)
}

func (p *HandlerParser) primitiveParam(call *ast.CallExpr, in string) *spec.Parameter {
	if len(call.Args) <= 0 {
		return nil
	}
	arg0 := call.Args[0]
	arg0Lit, ok := arg0.(*ast.BasicLit)
	if !ok {
		return nil
	}
	name := strings.Trim(arg0Lit.Value, "\"")
	res := &spec.Parameter{ParamProps: spec.ParamProps{Name: name, In: in}}

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
	param.Type = "array"
	param.Items = &spec.Items{SimpleSchema: spec.SimpleSchema{Type: "string"}}
	p.spec.AddParam(param)
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
				if !lo.Contains(p.spec.Produces, contentType) {
					p.spec.Produces = append(p.spec.Produces, contentType)
				}

				res := spec.NewResponse()
				commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
				if commentGroup != nil {
					comment := analyzer.ParseComment(commentGroup)
					if comment != nil {
						res.Description = comment.Text
					}
				}

				res.Schema = p.parseDataType(call, rule.Return.Data, contentType)
				statusCode := p.parseStatusCodeInCall(call, rule.Return.Status)
				p.spec.RespondsWith(statusCode, res)
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

				contentType := p.getRequestContentType(rule.Return.ContentType)
				switch contentType {
				case analyzer.MimeTypeFormData, analyzer.MimeTypeFormUrlencoded:
					in := p.getFormDataIn()
					params := p.parseParamsInCall(call, rule.Return.Data, contentType)
					for _, param := range params {
						param.In = in
						p.spec.AddParam(param)
					}

				default:
					schema := p.parseDataType(call, rule.Return.Data, contentType)
					if schema == nil {
						return
					}
					param := spec.BodyParam("payload", schema)

					commentGroup := p.ctx.GetHeadingCommentOf(call.Pos())
					if commentGroup != nil {
						comment := analyzer.ParseComment(commentGroup)
						if comment != nil {
							param.Description = comment.Text
						}
					}

					p.spec.AddParam(param)
				}

			},
		)
	}

	return
}

func (p *HandlerParser) getFormDataIn() string {
	var in = "query"
	switch p.api.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
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

func (p *HandlerParser) parseDataType(call *ast.CallExpr, dataType *DataSchema, contentType string) (schema *spec.Schema) {
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
		schema := p.basicSchemaType("array")
		schema.Items = &spec.SchemaOrArray{}
		schema.Items.Schema = p.parseDataType(call, dataType.Item, contentType)
		return schema
	case DataTypeObject:
		schema = p.basicSchemaType("object")
		schema.Properties = make(spec.SchemaProperties)
		for name, dataSchema := range dataType.Properties {
			s := p.parseDataType(call, dataSchema, contentType)
			if s != nil {
				schema.Properties[name] = *s
			}
		}
		return schema
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

func (p *HandlerParser) basicSchemaType(t string) *spec.Schema {
	return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{t}}}
}

func (p *HandlerParser) parseParamsInCall(call *ast.CallExpr, dataType *DataSchema, contentType string) (params []*spec.Parameter) {
	switch dataType.Type {
	case DataTypeString, DataTypeNumber, DataTypeInteger, DataTypeBoolean, DataTypeFile, DataTypeArray:
		param := &spec.Parameter{}
		param.Type = string(dataType.Type)
		param.Format = dataType.Format
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
