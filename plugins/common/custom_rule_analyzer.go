package common

import (
	"fmt"
	"go/ast"
	"log"
	"net/http"
	"strings"

	analyzer "github.com/gotomicro/eapi"
	"github.com/gotomicro/eapi/spec"
	"github.com/robertkrimen/otto"
	"github.com/samber/lo"
)

type CustomRuleAnalyzer struct {
	ctx  *analyzer.Context
	spec *analyzer.APISpec
	api  *analyzer.API
	c    *Config
}

func NewCustomRuleAnalyzer(ctx *analyzer.Context, spec *analyzer.APISpec, api *analyzer.API, c *Config) *CustomRuleAnalyzer {
	return &CustomRuleAnalyzer{ctx: ctx, spec: spec, api: api, c: c}
}

func (p *CustomRuleAnalyzer) MatchCustomResponseRule(node ast.Node) (matched bool) {
	if p.c == nil || len(p.c.Response) == 0 {
		return false
	}

	for idx, rule := range p.c.Response {
		var responseTypeTitle = "CustomResponseType"
		if idx > 0 {
			responseTypeTitle = fmt.Sprintf("CustomResponseType%d", idx)
		}

		genericTypeRef := spec.RefComponentSchemas(responseTypeTitle)
		genericType, ok := p.ctx.Doc().Components.Schemas[responseTypeTitle]
		if !ok {
			genericType = NewDataSchemaTransformer(rule.Return.Data).TransformToGeneric()
			genericType.Title = responseTypeTitle
			p.ctx.Doc().Components.Schemas[responseTypeTitle] = genericType
		}

		p.ctx.MatchCall(
			node,
			analyzer.NewCallRule().WithRule(rule.Type, rule.Method),
			func(call *ast.CallExpr, typeName, fnName string) {
				matched = true
				var contentType = rule.Return.ContentType
				res := spec.NewResponse()
				comment := p.ctx.ParseComment(p.ctx.GetHeadingCommentOf(call.Pos()))
				res.Description = comment.TextPointer()
				schema := p.parseDataType(call, rule.Return.Data, contentType, genericTypeRef)
				res.WithContent(spec.NewContentWithSchemaRef(schema, []string{contentType}))
				statusCode := p.parseStatusCodeInCall(call, rule.Return.Status)
				p.spec.AddResponse(statusCode, res)
			},
		)

		if genericType.IsTypeAlias() {
			delete(p.ctx.Doc().Components.Schemas, responseTypeTitle)
		}
	}

	return
}

func (p *CustomRuleAnalyzer) MatchCustomRequestRule(node ast.Node) (matched bool) {
	if p.c == nil || len(p.c.Request) == 0 {
		return false
	}

	for idx, rule := range p.c.Request {
		var requestTypeTitle = "CustomRequestType"
		if idx > 0 {
			requestTypeTitle = fmt.Sprintf("CustomRequestType%d", idx)
		}
		genericTypeRef := spec.RefComponentSchemas(requestTypeTitle)
		genericType, ok := p.ctx.Doc().Components.Schemas[requestTypeTitle]
		if !ok {
			genericType = NewDataSchemaTransformer(rule.Return.Data).TransformToGeneric()
			genericType.Title = requestTypeTitle
			p.ctx.Doc().Components.Schemas[requestTypeTitle] = genericType
		}

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
					schema := p.parseDataType(call, rule.Return.Data, contentType, genericTypeRef)
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
					p.spec.RequestBody = reqBody
				}

			},
		)

		if genericType.IsTypeAlias() {
			delete(p.ctx.Doc().Components.Schemas, requestTypeTitle)
		}
	}

	return
}

func (p *CustomRuleAnalyzer) parseDataType(call *ast.CallExpr, dataType *DataSchema, contentType string, genericType *spec.SchemaRef) (schema *spec.SchemaRef) {
	res := NewDataSchemaTransformer(dataType).TransformToSpecific(genericType, func(dataType *DataSchema) *spec.SchemaRef {
		output := p.evaluate(call, string(dataType.Type))
		if output == nil {
			return spec.NewObjectSchema().WithExtendedType(spec.NewNullExtType()).NewRef()
		}
		expr, ok := output.(ast.Expr)
		if !ok {
			fmt.Printf("invalid data type '%s' in configuration file\n", dataType.Type)
			return nil
		}
		return p.ctx.GetSchemaByExpr(expr, contentType)
	})

	resUnref := spec.Unref(p.ctx.Doc(), res)
	ext := resUnref.ExtendedTypeInfo
	if ext != nil && ext.Type == spec.ExtendedTypeSpecific {
		generic := spec.Unref(p.ctx.Doc(), ext.SpecificType.Type)
		if len(ext.SpecificType.Args) == 1 && generic.IsTypeAlias() {
			return ext.SpecificType.Args[0]
		}
	}

	return res
}

func (p *CustomRuleAnalyzer) parseParamsInCall(call *ast.CallExpr, dataType *DataSchema, contentType string) (params []*spec.Parameter) {
	switch dataType.Type {
	case DataTypeString, DataTypeNumber, DataTypeInteger, DataTypeBoolean, DataTypeFile, DataTypeArray:
		param := &spec.Parameter{}
		schema := spec.NewSchema()
		schema.Type = string(dataType.Type)
		schema.Format = dataType.Format
		param.Schema = schema
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
		return analyzer.NewParamParser(p.ctx, p.paramNameParser).Parse(expr)
	}
}

// 获取一个尽可能正确的 request payload contentType
func (p *CustomRuleAnalyzer) getRequestContentType(contentType string) string {
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

func (p *CustomRuleAnalyzer) paramNameParser(fieldName string, tags map[string]string) (name, in string) {
	name, ok := tags["form"]
	if ok {
		name, _, _ = strings.Cut(name, ",")
		return name, "query"
	}
	return fieldName, "query"
}

func (p *CustomRuleAnalyzer) parseStatusCodeInCall(call *ast.CallExpr, statusCode string) (code int) {
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

func (p *CustomRuleAnalyzer) evaluate(call *ast.CallExpr, code string) interface{} {
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
