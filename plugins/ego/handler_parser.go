package ego

import (
	"go/ast"
	"strings"

	"ego-gen-api"

	"github.com/go-openapi/spec"
)

const ginContextIdentName = "*github.com/gin-gonic/gin.Context"

var (
	interestedGinContextMethods = []string{"Bind", "JSON", "Query", "Param"}
)

type handlerParser struct {
	ctx  *analyzer.Context
	spec *analyzer.APISpec
	decl *ast.FuncDecl
}

func newHandlerParser(ctx *analyzer.Context, spec *analyzer.APISpec, decl *ast.FuncDecl) *handlerParser {
	return &handlerParser{ctx: ctx, spec: spec, decl: decl}
}

func (p *handlerParser) Parse() {
	ast.Inspect(p.decl, func(node ast.Node) bool {
		p.ctx.MatchCall(node,
			analyzer.NewCallRule().WithRule(ginContextIdentName, interestedGinContextMethods...),
			func(call *ast.CallExpr, typeName, fnName string) {
				switch fnName {
				case "Bind":
					p.parseBinding(call)
				case "JSON":
					p.parseJsonRes(call)
				case "Query": // query parameter
					p.parsePrimitiveParam(call, "query")
				case "Param": // path parameter
					p.parsePrimitiveParam(call, "path")
				}
			},
		)
		return true
	})
}

func (p *handlerParser) parseBinding(call *ast.CallExpr) {
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

func (p *handlerParser) parseJsonRes(call *ast.CallExpr) {
	if len(call.Args) != 2 {
		return
	}
	var contentType string
	if len(p.spec.Produces) > 0 {
		contentType = p.spec.Produces[0]
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

func (p *handlerParser) parsePrimitiveParam(call *ast.CallExpr, in string) {
	if len(call.Args) <= 0 {
		return
	}
	arg0 := call.Args[0]
	arg0Lit, ok := arg0.(*ast.BasicLit)
	if !ok {
		return
	}

	name := strings.Trim(arg0Lit.Value, "\"")
	param := &spec.Parameter{ParamProps: spec.ParamProps{Name: name, In: in}}
	p.spec.AddParam(param)
}
