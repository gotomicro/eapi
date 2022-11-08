package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"github.com/go-openapi/spec"
	"golang.org/x/tools/go/packages"
)

type RouteAnalyzer func(ctx *Context, node ast.Node) (routes []*API)

type Context struct {
	Env *Environment

	pkg      *packages.Package
	file     *ast.File
	analyzer *Analyzer
}

func newContext(analyzer *Analyzer, env *Environment) *Context {
	return &Context{
		Env:      env,
		analyzer: analyzer,
	}
}

func (c *Context) WithPackage(pkg *packages.Package) *Context {
	res := *c
	res.pkg = pkg
	return &res
}

func (c *Context) WithFile(file *ast.File) *Context {
	res := *c
	res.file = file
	return &res
}

func (c *Context) Block() *Context {
	res := *c
	res.Env = NewEnvironment(c.Env)
	return &res
}

func (c *Context) Package() *packages.Package {
	return c.pkg
}

func (c *Context) File() *ast.File {
	return c.file
}

func (c *Context) LineColumn(pos token.Pos) string {
	return c.pkg.Fset.Position(pos).String()
}

func (c *Context) GetDefinition(pkg, name string) Definition {
	return c.analyzer.definitions[pkg+"."+name]
}

func (c *Context) ParseType(t types.Type) Definition {
	switch t := t.(type) {
	case *types.Pointer:
		return c.ParseType(t.Elem())
	case *types.Named:
		return c.GetDefinition(t.Obj().Pkg().Path(), t.Obj().Name())
	case types.Object:
		return c.GetDefinition(t.Pkg().Path(), t.Name())
	}

	return nil
}

func (c *Context) Doc() *spec.Swagger {
	return c.analyzer.Doc()
}

func (c *Context) AddAPI(items ...*API) {
	c.analyzer.AddRoutes(items...)
}

func (c *Context) ParseStatusCode(status ast.Expr) int {
	switch status := status.(type) {
	case *ast.SelectorExpr:
		return c.ParseStatusCode(status.Sel)
	case *ast.Ident:
		obj := c.pkg.TypesInfo.Uses[status]
		if obj == nil {
			break
		}
		objConst, ok := obj.(*types.Const)
		if !ok {
			break
		}
		statusCode, err := strconv.ParseInt(objConst.Val().String(), 10, 64)
		if err != nil {
			break
		}
		return int(statusCode)

	case *ast.BasicLit:
	default:
		// unknown status code
		fmt.Printf("unknown status code %s\n", c.LineColumn(status.Pos()))
	}

	// unknown status code
	fmt.Printf("unknown status code %s\n", c.LineColumn(status.Pos()))

	// fallback to 200
	return 200
}

func (c *Context) GetSchemaByExpr(expr ast.Expr, contentType string) *spec.Schema {
	return NewSchemaBuilder(c, contentType).GetSchemaByExpr(expr, contentType)
}

func (c *Context) FindHeadCommentOf(pos token.Pos) *ast.CommentGroup {
	if c.File() == nil {
		return nil
	}

	position := c.Package().Fset.Position(pos)
	for _, commentGroup := range c.File().Comments {
		commentPos := c.Package().Fset.Position(commentGroup.End())
		if commentPos.Line == position.Line-1 {
			return commentGroup
		}
	}

	return nil
}

func (c *Context) APIs() *APIs {
	return c.analyzer.APIs()
}

func (c *Context) NewEnv() *Context {
	res := *c
	res.Env = NewEnvironment(nil)
	return &res
}

type CallRule struct {
	Rules map[string][]string // typeName to function-names
}

func NewCallRule() *CallRule {
	r := new(CallRule)
	r.Rules = make(map[string][]string)
	return r
}

func (c *CallRule) WithRule(typeName string, fnNames ...string) *CallRule {
	c.Rules[typeName] = append(c.Rules[typeName], fnNames...)
	return c
}

func (c *Context) MatchCall(n ast.Node, rule *CallRule, callback func(call *ast.CallExpr, typeName, fnName string)) {
	callExpr, ok := n.(*ast.CallExpr)
	if !ok {
		return
	}

	actualTypeName, actualFnName, err := c.GetCallInfo(callExpr)
	if err != nil {
		return
	}

	for typeName, fnNames := range rule.Rules {
		for _, fnName := range fnNames {
			if typeName == actualTypeName && fnName == actualFnName {
				callback(callExpr, typeName, fnName)
			}
		}
	}

	return
}

// GetCallInfo returns the package or type and name associated with a call expression
//
// e.g. GetCallInfo(`c.GET("/ping", ...)`) returns ("*github/gin-gonic/gin.RouterGroup", "GET", nil)
//
func (c *Context) GetCallInfo(n ast.Node) (string, string, error) {
	switch node := n.(type) {
	case *ast.CallExpr:
		switch fn := node.Fun.(type) {
		case *ast.SelectorExpr:
			switch expr := fn.X.(type) {
			case *ast.Ident:
				if expr.Obj != nil && expr.Obj.Kind == ast.Var {
					t := c.Package().TypesInfo.TypeOf(expr)
					if t != nil {
						return t.String(), fn.Sel.Name, nil
					}
					return "undefined", fn.Sel.Name, fmt.Errorf("missing type info")
				}
				return expr.Name, fn.Sel.Name, nil
			case *ast.SelectorExpr:
				if expr.Sel != nil {
					t := c.Package().TypesInfo.TypeOf(expr.Sel)
					if t != nil {
						return t.String(), fn.Sel.Name, nil
					}
					return "undefined", fn.Sel.Name, fmt.Errorf("missing type info")
				}
			case *ast.CallExpr:
				switch call := expr.Fun.(type) {
				case *ast.Ident:
					if call.Name == "new" {
						t := c.Package().TypesInfo.TypeOf(expr.Args[0])
						if t != nil {
							return t.String(), fn.Sel.Name, nil
						}
						return "undefined", fn.Sel.Name, fmt.Errorf("missing type info")
					}
					if call.Obj != nil {
						switch decl := call.Obj.Decl.(type) {
						case *ast.FuncDecl:
							ret := decl.Type.Results
							if ret != nil && len(ret.List) > 0 {
								ret1 := ret.List[0]
								if ret1 != nil {
									t := c.Package().TypesInfo.TypeOf(ret1.Type)
									if t != nil {
										return t.String(), fn.Sel.Name, nil
									}
									return "undefined", fn.Sel.Name, fmt.Errorf("missing type info")
								}
							}
						}
					}
				}
			}
		case *ast.Ident:
			return c.Package().Name, fn.Name, nil
		}
	}

	return "", "", fmt.Errorf("unable to determine call info")
}
