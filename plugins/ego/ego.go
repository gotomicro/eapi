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
	routerGroupTypeName   = "*github.com/gin-gonic/gin.RouterGroup"
)

var _ analyzer.Plugin = &Plugin{}

type Plugin struct{}

func (e *Plugin) ParseRoutes(ctx *analyzer.Context, node ast.Node, routes []*analyzer.Route) []*analyzer.Route {
	switch n := node.(type) {
	case *ast.AssignStmt:
		e.assignStmt(ctx, n, routes)
	case *ast.CallExpr:
		routes = e.callExpr(ctx, n, routes)
	}
	return routes
}

func (e *Plugin) ParseHandler(ctx *analyzer.Context, fn *ast.FuncDecl, spec *analyzer.APISpec) {
	NewHandlerParser(ctx, spec, fn).Parse()
}

func (e *Plugin) assignStmt(ctx *analyzer.Context, node ast.Node, routes []*analyzer.Route) {
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

func (e *Plugin) callExpr(ctx *analyzer.Context, node ast.Node, routes []*analyzer.Route) (res []*analyzer.Route) {
	res = routes
	callExpr, ok := node.(*ast.CallExpr)
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
	if !lo.Contains(routeCallMethods, selExpr.Sel.String()) {
		return
	}
	if len(callExpr.Args) < 2 {
		return
	}
	arg0, ok := callExpr.Args[0].(*ast.BasicLit)
	if !ok {
		return
	}

	var prefix = ""
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

	fullPath := path.Join(prefix, e.normalizePath(strings.Trim(arg0.Value, "\"")))
	res = append(res, &analyzer.Route{
		Method:   selExpr.Sel.String(),
		FullPath: fullPath,
		Handler:  handlerFn,
	})
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
