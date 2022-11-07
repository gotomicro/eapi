package ego

import (
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"regexp"
	"strings"

	analyzer "github.com/gotomicro/ego-gen-api"
)

var (
	routeMethods = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}
)

const (
	ginRouterGroupTypeName = "*github.com/gin-gonic/gin.RouterGroup"
	eginComponentTypeName  = "*github.com/gotomicro/ego/server/egin.Component"
	routerGroupMethodName  = "Group"
)

var _ analyzer.Plugin = &Plugin{}

type Plugin struct{}

func NewPlugin() *Plugin {
	return &Plugin{}
}

func (e *Plugin) Analyze(ctx *analyzer.Context, node ast.Node) {
	switch n := node.(type) {
	case *ast.AssignStmt:
		e.assignStmt(ctx, n)
	case *ast.CallExpr:
		e.callExpr(ctx, n)
	}
}

func (e *Plugin) Name() string {
	return "ego"
}

func (e *Plugin) assignStmt(ctx *analyzer.Context, node ast.Node) {
	assign := node.(*ast.AssignStmt)
	if len(assign.Rhs) != 1 || len(assign.Lhs) != 1 {
		return
	}

	rh := assign.Rhs[0]
	ctx.MatchCall(
		rh,
		analyzer.NewCallRule().
			WithRule(eginComponentTypeName, routerGroupMethodName).
			WithRule(ginRouterGroupTypeName, routerGroupMethodName),
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
			if rg, ok := v.(*analyzer.RouteGroup); ok {
				prefix = rg.Prefix
			}

			rg := &analyzer.RouteGroup{Prefix: path.Join(prefix, e.normalizePath(strings.Trim(arg0.Value, "\"")))}
			lh := assign.Lhs[0]
			lhIdent, ok := lh.(*ast.Ident)
			if !ok {
				return
			}

			switch assign.Tok {
			case token.ASSIGN:
				ctx.Env.Assign(lhIdent.Name, rg)
			case token.DEFINE:
				ctx.Env.Define(lhIdent.Name, rg)
			}
		},
	)

	return
}

func (e *Plugin) callExpr(ctx *analyzer.Context, callExpr *ast.CallExpr) {
	ctx.MatchCall(
		callExpr,
		analyzer.NewCallRule().
			WithRule(ginRouterGroupTypeName, routeMethods...).
			WithRule(eginComponentTypeName, routeMethods...),
		func(call *ast.CallExpr, typeName, fnName string) {
			api := e.parseAPI(ctx, callExpr)
			if api == nil {
				return
			}
			ctx.AddAPI(api)
		},
	)
}

func (e *Plugin) parseAPI(ctx *analyzer.Context, callExpr *ast.CallExpr) (api *analyzer.API) {
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
		if rg, ok := v.(*analyzer.RouteGroup); ok {
			prefix = rg.Prefix
		}
	}

	handlerArg := callExpr.Args[len(callExpr.Args)-1]
	var handler interface{}
	switch handlerArg := handlerArg.(type) {
	case *ast.Ident:
		handler = ctx.Package().TypesInfo.Uses[handlerArg]
	case *ast.SelectorExpr:
		handler = ctx.Package().TypesInfo.Uses[handlerArg.Sel]
	default:
		return // ignore
	}
	handlerFn, ok := handler.(*types.Func)
	if !ok {
		return
	}
	handlerDef := ctx.GetDefinition(handlerFn.Pkg().Path(), handlerFn.Name())
	handlerFnDef, ok := handlerDef.(*analyzer.FuncDefinition)
	if !ok {
		return
	}

	fullPath := path.Join(prefix, e.normalizePath(strings.Trim(arg0.Value, "\"")))
	method := selExpr.Sel.Name
	api = analyzer.NewAPI(method, fullPath)
	newHandlerParser(
		ctx.NewEnv().WithPackage(handlerFnDef.Pkg()).WithFile(handlerFnDef.File()),
		api.Spec,
		handlerFnDef.Decl,
	).Parse()
	api.Spec.LoadFromFuncDecl(handlerFnDef.Decl)
	return
}

var (
	pathParamPattern = regexp.MustCompile(":([^\\/]+)")
)

func (e *Plugin) normalizePath(path string) string {
	return pathParamPattern.ReplaceAllStringFunc(path, func(s string) string {
		s = strings.TrimPrefix(s, ":")
		return "{" + s + "}"
	})
}
