package gin

import (
	"go/ast"
	"strings"

	"github.com/go-openapi/spec"
	analyzer "github.com/gotomicro/ego-gen-api"
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
}

func NewHandlerParser(ctx *analyzer.Context, spec *analyzer.APISpec, decl *ast.FuncDecl) *HandlerParser {
	return &HandlerParser{ctx: ctx, spec: spec, decl: decl}
}

func (p *HandlerParser) Parse() {
	ast.Inspect(p.decl, func(node ast.Node) bool {
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
