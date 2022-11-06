package ego

import (
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"regexp"
	"strings"

	"ego-gen-api"

	"github.com/samber/lo"
)

var (
	routerIndents = []string{
		"*github.com/gin-gonic/gin.RouterGroup",
		"*github.com/gotomicro/ego/server/egin.Component",
	}
	routeCallMethods = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}
)

const (
	routerGroupMethodName = "Group"
)

var _ analyzer.Plugin = &Plugin{}

type Plugin struct{}

func (e *Plugin) Analyze(ctx *analyzer.Context, node ast.Node) {
	switch n := node.(type) {
	case *ast.AssignStmt:
		e.assignStmt(ctx, n)
	case *ast.CallExpr:
		e.callExpr(ctx, n)
	}
}

func (e *Plugin) assignStmt(ctx *analyzer.Context, node ast.Node) {
	assign := node.(*ast.AssignStmt)
	if len(assign.Rhs) != 1 || len(assign.Lhs) != 1 {
		return
	}

	rh := assign.Rhs[0]
	callExpr, ok := rh.(*ast.CallExpr)
	if !ok {
		return
	}
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	t := ctx.Package().TypesInfo.TypeOf(selExpr.X)
	if t == nil {
		return
	}
	if !lo.Contains(routerIndents, t.String()) {
		return
	}
	if selExpr.Sel.Name != routerGroupMethodName {
		return
	}
	if len(callExpr.Args) <= 0 {
		return
	}
	arg0, ok := callExpr.Args[0].(*ast.BasicLit)
	if !ok {
		return
	}
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

	return
}

func (e *Plugin) callExpr(ctx *analyzer.Context, callExpr *ast.CallExpr) {
	api := e.parseAPI(ctx, callExpr)
	if api == nil {
		return
	}

	ctx.AddAPI(api)
	return
}

func (e *Plugin) parseAPI(ctx *analyzer.Context, callExpr *ast.CallExpr) *analyzer.API {
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	t := ctx.Package().TypesInfo.TypeOf(selExpr.X)
	if t == nil {
		return nil
	}
	if !lo.Contains(routerIndents, t.String()) {
		return nil
	}
	if !lo.Contains(routeCallMethods, selExpr.Sel.String()) {
		return nil
	}
	if len(callExpr.Args) < 2 {
		return nil
	}
	arg0, ok := callExpr.Args[0].(*ast.BasicLit)
	if !ok {
		return nil
	}

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
		return nil // ignore
	}
	handlerFn, ok := handler.(*types.Func)
	if !ok {
		return nil
	}
	fullPath := path.Join(prefix, e.normalizePath(strings.Trim(arg0.Value, "\"")))
	method := selExpr.Sel.String()

	route := analyzer.NewAPI(method, fullPath)

	handlerDef := ctx.GetDefinition(handlerFn.Pkg().Path(), handlerFn.Name())
	handlerFnDef, ok := handlerDef.(*analyzer.FuncDefinition)
	if !ok {
		return nil
	}
	newHandlerParser(
		ctx.NewEnv().WithPackage(handlerFnDef.Pkg()).WithFile(handlerFnDef.File()),
		route.Spec,
		handlerFnDef.Decl,
	).Parse()
	route.Spec.LoadFromFuncDecl(handlerFnDef.Decl)
	return route
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
