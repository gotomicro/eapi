package gin

import (
	"go/ast"
	"log"
	"strings"

	"github.com/go-openapi/spec"
	analyzer "github.com/gotomicro/ego-gen-api"
	"github.com/robertkrimen/otto"
	"github.com/samber/lo"
)

const ginContextIdentName = "*github.com/gin-gonic/gin.Context"

var (
	interestedGinContextMethods = []string{"Bind", "JSON", "Query", "Param", "GetPostForm", "PostFormArray", "XML"}
)

type HandlerParser struct {
	ctx  *analyzer.Context
	spec *analyzer.APISpec
	decl *ast.FuncDecl

	c *Config
}

func NewHandlerParser(ctx *analyzer.Context, spec *analyzer.APISpec, decl *ast.FuncDecl) *HandlerParser {
	return &HandlerParser{ctx: ctx, spec: spec, decl: decl}
}

func (p *HandlerParser) WithConfig(c *Config) *HandlerParser {
	p.c = c
	return p
}

func (p *HandlerParser) Parse() {
	ast.Inspect(p.decl, func(node ast.Node) bool {
		matched := p.matchCustomRule(node)
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

	var contentType string
	if len(p.spec.Consumes) > 0 {
		contentType = p.spec.Consumes[0]
	}
	payloadSchema := p.ctx.GetSchemaByExpr(call.Args[0], contentType)
	p.spec.AddParam(spec.BodyParam("Payload", payloadSchema))
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
	commentGroup := p.ctx.FindHeadCommentOf(call.Pos())
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

	commentGroup := p.ctx.FindHeadCommentOf(call.Pos())
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

func (p *HandlerParser) matchCustomRule(node ast.Node) (matched bool) {
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
				commentGroup := p.ctx.FindHeadCommentOf(call.Pos())
				if commentGroup != nil {
					comment := analyzer.ParseComment(commentGroup)
					if comment != nil {
						res.Description = comment.Text
					}
				}

				res.Schema = p.parseSchemaInCall(call, rule.Return.Data, contentType)
				statusCode := p.parseStatusCodeInCall(call, rule.Return.Status)
				p.spec.RespondsWith(statusCode, res)
			},
		)
	}

	return
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

func (p *HandlerParser) parseSchemaInCall(call *ast.CallExpr, code string, contentType string) (schema *spec.Schema) {
	output := p.evaluate(call, code)
	expr, ok := output.(ast.Expr)
	if !ok {
		return nil
	}
	schema = p.ctx.GetSchemaByExpr(expr, contentType)

	return
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
