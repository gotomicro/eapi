package gin

import (
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"regexp"
	"strings"

	analyzer "github.com/gotomicro/ego-gen-api"
	"github.com/spf13/viper"
)

var (
	routeMethods = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}
)

const (
	ginRouterGroupTypeName = "*github.com/gin-gonic/gin.RouterGroup"
	ginIRouterTypeName     = "github.com/gin-gonic/gin.IRouter"
	ginIRoutesTypeName     = "github.com/gin-gonic/gin.IRoutes"
	routerGroupMethodName  = "Group"
)

var _ analyzer.Plugin = &Plugin{}

type Plugin struct {
	config Config
}

func NewPlugin() *Plugin {
	return &Plugin{}
}

func (e *Plugin) Mount() error {
	err := viper.UnmarshalKey("properties", &e.config)
	if err != nil {
		return err
	}

	return nil
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
	return "gin"
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
			WithRule(ginRouterGroupTypeName, routerGroupMethodName).
			WithRule(ginIRouterTypeName, routerGroupMethodName).
			WithRule(ginIRoutesTypeName, routerGroupMethodName),
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

	return
}

func (e *Plugin) callExpr(ctx *analyzer.Context, callExpr *ast.CallExpr) {
	ctx.MatchCall(
		callExpr,
		analyzer.NewCallRule().WithRule(ginRouterGroupTypeName, routeMethods...).
			WithRule(ginIRouterTypeName, routeMethods...).
			WithRule(ginIRoutesTypeName, routeMethods...),
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

	handlerFn := e.getHandlerFn(ctx, callExpr)
	if handlerFn == nil {
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
	api.Spec.LoadFromFuncDecl(handlerFnDef.Decl)
	if api.Spec.ID == "" {
		api.Spec.ID = handlerFnDef.Pkg().Name + "." + handlerFnDef.Decl.Name.Name
	}
	NewHandlerParser(
		ctx.NewEnv().WithPackage(handlerFnDef.Pkg()).WithFile(handlerFnDef.File()),
		api,
		handlerFnDef.Decl,
	).WithConfig(&e.config).Parse()
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

func (e *Plugin) getHandlerFn(ctx *analyzer.Context, callExpr *ast.CallExpr) (handlerFn *types.Func) {
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
		return
	}

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
