package echo

import (
	"go/ast"
	"log"
	"net/http"
	"strings"

	analyzer "github.com/gotomicro/eapi"
	"github.com/gotomicro/eapi/plugins/common"
	"github.com/gotomicro/eapi/spec"
	"github.com/iancoleman/strcase"
	"github.com/robertkrimen/otto"
)

const echoContextIdentName = "github.com/labstack/echo/v4.Context"

var (
	interestedEchoContextMethods = []string{"Bind", "JSON", "QueryParam", "Param", "FormValue", "XML", "XMLPretty", "Redirect", "FormFile"}
)

type handlerAnalyzer struct {
	ctx  *analyzer.Context
	api  *analyzer.API
	spec *analyzer.APISpec
	decl *ast.FuncDecl

	c *common.Config
}

func newHandlerAnalyzer(ctx *analyzer.Context, api *analyzer.API, decl *ast.FuncDecl) *handlerAnalyzer {
	return &handlerAnalyzer{ctx: ctx, api: api, spec: api.Spec, decl: decl}
}

func (p *handlerAnalyzer) WithConfig(c *common.Config) *handlerAnalyzer {
	p.c = c
	return p
}

func (p *handlerAnalyzer) Parse() {
	ast.Inspect(p.decl, func(node ast.Node) bool {
		customRuleAnalyzer := common.NewCustomRuleAnalyzer(
			p.ctx,
			p.spec,
			p.api,
			p.c,
		)
		matched := customRuleAnalyzer.MatchCustomResponseRule(node)
		if matched {
			return true
		}
		matched = customRuleAnalyzer.MatchCustomRequestRule(node)
		if matched {
			return true
		}

		p.ctx.MatchCall(node,
			analyzer.NewCallRule().WithRule(echoContextIdentName, interestedEchoContextMethods...),
			func(call *ast.CallExpr, typeName, fnName string) {
				switch fnName {
				case "Bind":
					p.parseBinding(call)
				case "JSON":
					p.parseResBody(call, "application/json")
				case "XML", "XMLPretty":
					p.parseResBody(call, "application/xml")
				case "QueryParam": // query parameter
					p.parsePrimitiveParam(call, "query")
				case "Param": // path parameter
					p.parsePrimitiveParam(call, "path")
				case "FormValue":
					p.parseFormData(call, "string")
				case "FormFile":
					p.parseFormData(call, "file")
				case "Redirect":
					p.parseRedirectRes(call)
					// TODO: supporting more methods (FileForm(), HTML(), Data(), etc...)
				}
			},
		)
		return true
	})
}

func (p *handlerAnalyzer) paramNameParser(field string, tags map[string]string) (name string, in string) {
	name, ok := tags["query"]
	if ok {
		name, _, _ = strings.Cut(name, ",")
		return name, "query"
	}
	// fallback
	return field, "query"
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
		contentType := p.getDefaultContentType()
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
		p.spec.RequestBody = reqBody
	}
}

func (p *handlerAnalyzer) parseResBody(call *ast.CallExpr, contentType string) {
	if len(call.Args) < 2 {
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
		requestBody = spec.NewRequestBody().WithContent(spec.NewContent())
		p.spec.RequestBody = requestBody
	}
	mediaType := requestBody.GetMediaType(analyzer.MimeTypeFormData)
	if mediaType == nil {
		mediaType = spec.NewMediaType()
		requestBody.Content[analyzer.MimeTypeFormData] = mediaType
	}

	comment := p.ctx.ParseComment(p.ctx.GetHeadingCommentOf(call.Pos()))
	paramSchema.Description = comment.Text()

	var schemaRef = mediaType.Schema
	var schema *spec.SchemaRef
	if schemaRef != nil {
		schema = spec.Unref(p.ctx.Doc(), schemaRef)
		schema.WithProperty(name, paramSchema)
	} else {
		schema = spec.NewObjectSchema().NewRef()
		title := strcase.ToCamel(p.spec.OperationID) + "Request"
		schema.Title = title
		schema.WithProperty(name, paramSchema)
		p.ctx.Doc().Components.Schemas[title] = schema
		schemaRef = spec.RefComponentSchemas(title)
		mediaType.Schema = schemaRef
	}
	if comment.Required() {
		schema.Required = append(schema.Required, name)
	}
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

	var res *spec.Parameter
	switch in {
	case "path":
		res = spec.NewPathParameter(name).WithSchema(paramSchema)
	case "query":
		res = spec.NewQueryParameter(name).WithSchema(paramSchema)
	}

	comment := p.ctx.ParseComment(p.ctx.GetHeadingCommentOf(call.Pos()))
	res.Description = comment.Text()

	return res
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
func (p *handlerAnalyzer) getDefaultContentType() string {
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
