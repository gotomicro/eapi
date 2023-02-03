package echo

import (
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"regexp"
	"strings"

	"github.com/gotomicro/eapi"
	"github.com/gotomicro/eapi/plugins/common"
	"github.com/knadh/koanf"
)

var (
	routeMethods = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}
)

const (
	echoInstanceTypeName = "*github.com/labstack/echo/v4.Echo"
	echoGroupTypeName    = "*github.com/labstack/echo/v4.Group"
	echoGroupMethodName  = "Group"
)

type Plugin struct {
	config common.Config
}

func NewPlugin() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "echo"
}

func (p *Plugin) Mount(k *koanf.Koanf) error {
	return k.Unmarshal("properties", &p.config)
}

func (p *Plugin) Analyze(ctx *eapi.Context, node ast.Node) {
	switch node := node.(type) {
	case *ast.AssignStmt:
		p.assignStmt(ctx, node)

	case *ast.CallExpr:
		p.callExpr(ctx, node)
	}
}

func (p *Plugin) assignStmt(ctx *eapi.Context, assign *ast.AssignStmt) {
	if len(assign.Rhs) != 1 || len(assign.Lhs) != 1 {
		return
	}

	callRule := eapi.NewCallRule().
		WithRule(echoInstanceTypeName, echoGroupMethodName).
		WithRule(echoGroupTypeName, echoGroupMethodName)
	for _, router := range p.config.RouterNames {
		callRule = callRule.WithRule(router, echoGroupMethodName)
	}

	rh := assign.Rhs[0]
	ctx.MatchCall(
		rh,
		callRule,
		func(callExpr *ast.CallExpr, typeName, fnName string) {
			if len(callExpr.Args) <= 0 {
				return
			}
			arg0, ok := callExpr.Args[0].(*ast.BasicLit)
			if !ok {
				return
			}
			selExpr := callExpr.Fun.(*ast.SelectorExpr)
			xIdent, ok := selExpr.X.(*ast.Ident)
			if !ok {
				return
			}
			var prefix = ""
			v := ctx.Env.Lookup(xIdent.Name)
			if rg, ok := v.(*eapi.RouteGroup); ok {
				prefix = rg.Prefix
			}
			rg := &eapi.RouteGroup{Prefix: path.Join(prefix, p.normalizePath(strings.Trim(arg0.Value, "\"")))}
			lh := assign.Lhs[0]
			lhIdent, ok := lh.(*ast.Ident)
			if !ok {
				return
			}
			switch assign.Tok {
			case token.ASSIGN:
				env := ctx.Env.Resolve(lhIdent.Name)
				if env == nil {
					ctx.Env.Define(lhIdent.Name, rg)
				} else {
					env.Assign(lhIdent.Name, rg)
				}

			case token.DEFINE:
				ctx.Env.Define(lhIdent.Name, rg)
			}
		},
	)

}

func (p *Plugin) callExpr(ctx *eapi.Context, callExpr *ast.CallExpr) {
	callRule := eapi.NewCallRule().WithRule(echoInstanceTypeName, routeMethods...).
		WithRule(echoGroupTypeName, routeMethods...)
	for _, router := range p.config.RouterNames {
		callRule = callRule.WithRule(router, routeMethods...)
	}

	ctx.MatchCall(
		callExpr,
		callRule,
		func(call *ast.CallExpr, typeName, fnName string) {
			comment := ctx.ParseComment(ctx.GetHeadingCommentOf(call.Lparen))
			if comment.Ignore() {
				return
			}
			api := p.parseAPI(ctx, callExpr)
			if api == nil {
				return
			}
			ctx.AddAPI(api)
		},
	)
}

func (e *Plugin) parseAPI(ctx *eapi.Context, callExpr *ast.CallExpr) (api *eapi.API) {
	if len(callExpr.Args) < 2 {
		return
	}
	arg0, ok := callExpr.Args[0].(*ast.BasicLit)
	if !ok {
		return
	}

	selExpr := callExpr.Fun.(*ast.SelectorExpr)
	var prefix string
	if xIdent, ok := selExpr.X.(*ast.Ident); ok {
		v := ctx.Env.Lookup(xIdent.Name)
		if rg, ok := v.(*eapi.RouteGroup); ok {
			prefix = rg.Prefix
		}
	}

	handlerFn := e.getHandlerFn(ctx, callExpr)
	if handlerFn == nil {
		return
	}

	handlerDef := ctx.GetDefinition(handlerFn.Pkg().Path(), handlerFn.Name())
	handlerFnDef, ok := handlerDef.(*eapi.FuncDefinition)
	if !ok {
		return
	}

	fullPath := path.Join(prefix, e.normalizePath(strings.Trim(arg0.Value, "\"")))
	method := selExpr.Sel.Name
	api = eapi.NewAPI(method, fullPath)
	api.Spec.LoadFromFuncDecl(ctx.Package().Fset, handlerFnDef.Decl)
	if api.Spec.OperationID == "" {
		api.Spec.OperationID = handlerFnDef.Pkg().Name + "." + handlerFnDef.Decl.Name.Name
	}
	if len(api.Spec.Tags) == 0 {
		api.Spec.Tags = ctx.Env.LookupTags()
	}
	newHandlerAnalyzer(
		ctx.NewEnv().WithPackage(handlerFnDef.Pkg()).WithFile(handlerFnDef.File()),
		api,
		handlerFnDef.Decl,
	).WithConfig(&e.config).Parse()

	return
}

func (e *Plugin) getHandlerFn(ctx *eapi.Context, callExpr *ast.CallExpr) (handlerFn *types.Func) {
	handlerArg := callExpr.Args[len(callExpr.Args)-1]

	if call, ok := handlerArg.(*ast.CallExpr); ok {
		nestedCall := e.unwrapCall(call)
		if len(nestedCall.Args) <= 0 {
			return
		}
		handlerArg = nestedCall.Args[0]
	}

	var handler interface{}
	switch handlerArg := handlerArg.(type) {
	case *ast.Ident:
		handler = ctx.Package().TypesInfo.Uses[handlerArg]
	case *ast.SelectorExpr:
		handler = ctx.Package().TypesInfo.Uses[handlerArg.Sel]
	default:
		return
	}

	handlerFn, ok := handler.(*types.Func)
	if !ok {
		return nil
	}
	return
}

// unwrap and returns the first nested call
// e.g. unwrapCall(`a(b(c(d)), b1(c1))`) return `c(d)`
func (e *Plugin) unwrapCall(callExpr *ast.CallExpr) *ast.CallExpr {
	if len(callExpr.Args) == 0 {
		return callExpr
	}

	arg0 := callExpr.Args[0]
	arg0Call, ok := arg0.(*ast.CallExpr)
	if ok {
		return e.unwrapCall(arg0Call)
	}

	return callExpr
}

var (
	pathParamPattern = regexp.MustCompile(`:([^\/]+)`)
)

func (e *Plugin) normalizePath(path string) string {
	return pathParamPattern.ReplaceAllStringFunc(path, func(s string) string {
		s = strings.TrimPrefix(s, ":")
		return "{" + s + "}"
	})
}
