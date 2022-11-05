package analyzer

import (
	"fmt"
	"go/ast"
	"go/build"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

type Analyzer struct {
	routes      []*Route
	globalEnv   *Environment
	plugins     []Plugin
	definitions Definitions

	doc      *spec.Swagger
	packages []*packages.Package
}

func NewAnalyzer() *Analyzer {
	a := &Analyzer{
		routes:      make([]*Route, 0),
		globalEnv:   NewEnvironment(nil),
		plugins:     make([]Plugin, 0),
		definitions: make(Definitions),
	}

	doc := &spec.Swagger{}
	doc.Info = &spec.Info{}
	doc.Swagger = "2.0"
	doc.Definitions = make(spec.Definitions)
	paths := spec.Paths{}
	paths.Paths = make(map[string]spec.PathItem)
	doc.Paths = &paths
	a.doc = doc

	return a
}

func (a *Analyzer) Plugin(plugins ...Plugin) *Analyzer {
	a.plugins = append(a.plugins, plugins...)
	return a
}

func (a *Analyzer) Load(packagePath string) {
	packagePath, err := filepath.Abs(packagePath)
	if err != nil {
		panic("invalid package path: " + err.Error())
	}
	a.packages = append(a.packages, a.load(packagePath)...)
}

func (a *Analyzer) Process(packagePath string) {
	if len(a.plugins) <= 0 {
		panic("must register plugin before processing")
	}

	a.Load(packagePath)

	for _, pkg := range a.packages {
		InspectPackage(pkg, func(pkg *packages.Package) bool {
			if pkg.Module != nil && pkg.Module.Dir == packagePath {
				a.processPkg(pkg)
				return true
			}
			return false
		})
	}

	for _, route := range a.routes {
		route.spec = &APISpec{spec.NewOperation(route.Method + " " + route.FullPath)}
		handler := route.Handler
		pkg, file, fnDecl := a.findFuncDeclInPackages(a.packages, handler.Pkg().Path(), handler.Name())
		if fnDecl == nil {
			continue
		}

		comment := ParseComment(fnDecl.Doc)
		if comment != nil {
			route.Operation().WithConsumes(comment.Consumes()...)
			route.Operation().WithProduces(comment.Produces()...)
			route.Operation().WithDescription(comment.TrimPrefix(fnDecl.Name.Name))
		}

		for _, plugin := range a.plugins {
			plugin.ParseHandler(a.context().withPackage(pkg).withFile(file), fnDecl, route.spec)
		}

		pathItem := a.doc.Paths.Paths[route.FullPath]
		route.applyToPathItem(&pathItem)
		a.doc.Paths.Paths[route.FullPath] = pathItem
	}
}

func (a *Analyzer) GetRoutes() []*Route {
	return a.routes
}

func (a *Analyzer) Doc() *spec.Swagger {
	return a.doc
}

func (a *Analyzer) parseRoutes(ctx *Context, node ast.Node) {
	for _, plugin := range a.plugins {
		a.routes = plugin.ParseRoutes(ctx, node, a.routes)
	}
}

func (a *Analyzer) load(pkgPath string) []*packages.Package {
	absPath, err := filepath.Abs(pkgPath)
	if err != nil {
		panic("invalid package path: " + pkgPath)
	}

	pkg, err := build.Default.ImportDir(absPath, build.ImportComment)
	if err != nil {
		panic("import directory failed: " + err.Error())
	}

	config := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedExportFile |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedModule |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes,
		BuildFlags: []string{},
		Tests:      false,
		Dir:        absPath,
	}
	var files []string
	for _, filename := range append(pkg.GoFiles, pkg.CgoFiles...) {
		files = append(files, filepath.Join(pkgPath, filename))
	}
	res, err := packages.Load(config, files...)
	if err != nil {
		panic("load packages failed: " + err.Error())
	}

	// 前面的 packages.Load() 方法不能解析出以第一层的 Module
	// 所以这里手动解析 go.mod
	for _, p := range res {
		p.Module = a.parseGoModule(pkgPath)
	}

	return res
}

func (a *Analyzer) processPkg(pkg *packages.Package) {
	ctx := a.context().Block().withPackage(pkg)
	for _, file := range pkg.Syntax {
		a.processFile(ctx.Block().withFile(file), file, pkg)
	}
}

func (a *Analyzer) processFile(ctx *Context, file *ast.File, pkg *packages.Package) {
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncDecl:
			a.funDecl(ctx, node, file, pkg)
			return false
		case *ast.TypeSpec:
			a.typeSpec(ctx, node, file, pkg)
			return false
		case *ast.BlockStmt:
			a.blockStmt(ctx, node, pkg)
			return false
		}

		a.parseRoutes(ctx, node)
		return true
	})
}

func (a *Analyzer) funDecl(ctx *Context, node *ast.FuncDecl, file *ast.File, pkg *packages.Package) {
	a.definitions.Set(NewFuncDefinition(pkg, file, node))

	ast.Inspect(node, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.BlockStmt:
			a.blockStmt(ctx, node, pkg)
			return false
		}

		//route := a.matchRoute(node, pkg, env)
		//if route != nil {
		//	a.routes = append(a.routes, route)
		//}

		a.parseRoutes(ctx, node)
		return true
	})
}

func (a *Analyzer) typeSpec(ctx *Context, node *ast.TypeSpec, file *ast.File, pkg *packages.Package) {
	a.definitions.Set(NewTypeDefinition(pkg, file, node))
}

func (a *Analyzer) blockStmt(ctx *Context, node *ast.BlockStmt, pkg *packages.Package) {
	for _, node := range node.List {
		ast.Inspect(node, func(node ast.Node) bool {
			a.parseRoutes(ctx, node)

			switch node := node.(type) {
			case *ast.BlockStmt:
				a.blockStmt(ctx.Block(), node, pkg)
				return false
			}
			return true
		})
	}
}

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

func (a *Analyzer) matchRoute(node ast.Node, pkg *packages.Package, env *Environment) (route *Route) {
	callExpr, ok := node.(*ast.CallExpr)
	if !ok {
		return
	}

	fnSel, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	xIdent, ok := fnSel.X.(*ast.Ident)
	if !ok {
		return
	}
	t := pkg.TypesInfo.TypeOf(xIdent)
	if t == nil {
		return
	}
	if !lo.Contains(routerIndents, t.String()) || !lo.Contains(routeCallMethods, fnSel.Sel.Name) {
		return
	}
	if len(callExpr.Args) <= 0 {
		return
	}

	var prefix string
	if t.String() == routerGroupTypeName {
		prefix = a.getGroupPrefix(xIdent, pkg)
	}

	arg0 := callExpr.Args[0]
	arg0Lit, ok := arg0.(*ast.BasicLit)
	if !ok {
		return
	}
	routePath := strings.Trim(arg0Lit.Value, "\"")
	route = &Route{
		Method:   fnSel.Sel.Name,
		FullPath: path.Join(prefix, routePath),
	}
	return
}

func (a *Analyzer) getGroupPrefix(routerGroupIdent *ast.Ident, pkg *packages.Package) (prefix string) {
	if routerGroupIdent.Obj.Kind != ast.Var {
		return
	}
	assignStmt, ok := routerGroupIdent.Obj.Decl.(*ast.AssignStmt)
	if !ok {
		return
	}

	if len(assignStmt.Lhs) != 1 || len(assignStmt.Rhs) != 1 { // 暂不支持多左值赋值表达式
		return
	}
	rhExpr := assignStmt.Rhs[0]
	callExpr, ok := rhExpr.(*ast.CallExpr)
	if !ok {
		return
	}
	fnSelExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	fnSelXIdent, ok := fnSelExpr.X.(*ast.Ident)
	if !ok {
		return
	}
	t := pkg.TypesInfo.TypeOf(fnSelXIdent)
	if t == nil {
		return
	}
	if !lo.Contains(routerIndents, t.String()) {
		return
	}
	if fnSelExpr.Sel.Name != routerGroupMethodName {
		return
	}

	arg0 := callExpr.Args[0]
	arg0Lit, ok := arg0.(*ast.BasicLit)
	if !ok {
		return
	}
	prefix = strings.Trim(arg0Lit.Value, "\"")

	if t.String() == routerGroupTypeName {
		prefix = path.Join(a.getGroupPrefix(fnSelXIdent, pkg), prefix)
	}

	return prefix
}

func (a *Analyzer) findFuncDeclInPackages(list []*packages.Package, pkgName, fnName string) (pkg *packages.Package, inFile *ast.File, decl *ast.FuncDecl) {
	for _, p := range list {
		pkg, inFile, decl = a.findFuncDeclInPackage(p, pkgName, fnName)
		if pkg != nil {
			return
		}
	}
	return
}

func (a *Analyzer) findFuncDeclInPackage(pkg *packages.Package, pkgName string, fnName string) (inPkg *packages.Package, inFile *ast.File, decl *ast.FuncDecl) {
	visited := make(map[string]bool)

	var visit func(p *packages.Package) bool
	visit = func(p *packages.Package) bool {
		if visited[p.PkgPath] {
			return true
		}
		visited[p.PkgPath] = true

		if p.PkgPath == pkgName {
			inPkg = p
			inspector.New(p.Syntax).WithStack([]ast.Node{&ast.FuncDecl{}}, func(n ast.Node, push bool, stack []ast.Node) bool {
				if !push {
					return false
				}
				fn := n.(*ast.FuncDecl)
				if fn.Name.Name == fnName {
					decl = fn
					inFile = stack[0].(*ast.File)
					return false
				}
				return true
			})
			return false
		}

		for _, p := range lo.Values(p.Imports) {
			if !visit(p) {
				return false
			}
		}
		return true
	}

	visit(pkg)
	return
}

func (a *Analyzer) parseGoModule(pkgPath string) *packages.Module {
	fileName := filepath.Join(pkgPath, "go.mod")
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		panic(err)
	}

	mod, err := modfile.Parse("go.mod", content, nil)
	if err != nil {
		panic(fmt.Sprintf("parse go.mod failed. %s. err=%s", fileName, err.Error()))
	}

	return &packages.Module{
		Path:      mod.Module.Mod.Path,
		Main:      true,
		Dir:       pkgPath,
		GoMod:     fileName,
		GoVersion: mod.Go.Version,
	}
}

func (a *Analyzer) context() *Context {
	return newContext(a, a.globalEnv)
}
