package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"github.com/go-openapi/spec"
	"golang.org/x/tools/go/packages"
)

type RouteAnalyzer func(ctx *Context, node ast.Node) (routes []*Route)

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

func (c *Context) withPackage(pkg *packages.Package) *Context {
	res := *c
	res.pkg = pkg
	return &res
}

func (c *Context) withFile(file *ast.File) *Context {
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
		fmt.Printf("unknown status code %s", c.LineColumn(status.Pos()))
	}

	// unknown status code
	fmt.Printf("unknown status code %s", c.LineColumn(status.Pos()))

	// fallback to 200
	return 200
}

func (c *Context) GetSchemaByExpr(expr ast.Expr, contentType string) *spec.Schema {
	t := c.Package().TypesInfo.TypeOf(expr)
	def := c.ParseType(t)
	typeDef, ok := def.(*TypeDefinition)
	if !ok {
		return nil
	}
	_, ok = c.Doc().Definitions[typeDef.ModelKey()]
	if ok {
		return spec.RefSchema(typeDef.RefKey())
	}

	payloadSchema := typeDef.ToSwaggerSchema(c, contentType)
	payloadSchema.ID = strings.ReplaceAll(typeDef.Key(), "/", "_")
	c.Doc().Definitions[typeDef.ModelKey()] = *payloadSchema
	return spec.RefSchema(typeDef.RefKey())
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
