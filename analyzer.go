package analyzer

import (
	"fmt"
	"go/ast"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotomicro/ego-gen-api/spec"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

type Analyzer struct {
	routes      APIs
	globalEnv   *Environment
	plugins     []Plugin
	definitions Definitions
	depends     []string

	doc      *spec.T
	packages []*packages.Package
}

func NewAnalyzer() *Analyzer {
	a := &Analyzer{
		routes:      make(APIs, 0),
		globalEnv:   NewEnvironment(nil),
		plugins:     make([]Plugin, 0),
		definitions: make(Definitions),
	}

	components := spec.NewComponents()
	components.Schemas = make(spec.Schemas)
	doc := &spec.T{
		OpenAPI:    "3.0.3",
		Info:       &spec.Info{},
		Components: components,
		Paths:      make(spec.Paths),
	}
	a.doc = doc

	return a
}

func (a *Analyzer) Plugin(plugins ...Plugin) *Analyzer {
	for _, plugin := range plugins {
		err := plugin.Mount()
		if err != nil {
			panic(fmt.Sprintf("mount plugin '%s' failed. error: %s", plugin.Name(), err.Error()))
		}
	}

	a.plugins = append(a.plugins, plugins...)
	return a
}

func (a *Analyzer) Depends(pkgNames ...string) *Analyzer {
	a.depends = append(a.depends, pkgNames...)
	return a
}

func (a *Analyzer) Load(packagePath string) {
	packagePath, err := filepath.Abs(packagePath)
	if err != nil {
		panic("invalid package path: " + err.Error())
	}

	pkgList := a.load(packagePath)
	for _, pkg := range pkgList {
		a.loadDefinitionsFromPkg(pkg, pkg.Module.Dir)
	}

	a.packages = append(a.packages, pkgList...)
}

func (a *Analyzer) Process(packagePath string) *Analyzer {
	if len(a.plugins) <= 0 {
		panic("must register plugin before processing")
	}

	packagePath, err := filepath.Abs(packagePath)
	if err != nil {
		panic("invalid package path: " + err.Error())
	}

	pkgList := a.load(packagePath)
	for _, pkg := range pkgList {
		a.loadDefinitionsFromPkg(pkg, pkg.Module.Dir)
	}
	a.packages = append(a.packages, pkgList...)

	a.processPkg(packagePath)

	return a
}

func (a *Analyzer) APIs() *APIs {
	return &a.routes
}

func (a *Analyzer) Doc() *spec.T {
	return a.doc
}

func (a *Analyzer) analyze(ctx *Context, node ast.Node) {
	for _, plugin := range a.plugins {
		plugin.Analyze(ctx, node)
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
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedModule |
			packages.NeedTypesInfo |
			0,
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
		if p.Module != nil {
			continue
		}

		module := a.parseGoModule(pkgPath)
		if module == nil {
			panic("failed to parse go.mod file in " + pkgPath)
		}
		p.Module = module
		p.PkgPath = module.Path
		p.ID = module.Path
	}

	return res
}

func (a *Analyzer) processPkg(packagePath string) {
	for _, pkg := range a.packages {
		moduleDir := pkg.Module.Dir
		InspectPackage(pkg, func(pkg *packages.Package) bool {
			if pkg.Module == nil || pkg.Module.Dir != moduleDir {
				return false
			}

			ctx := a.context().Block().WithPackage(pkg)
			for _, file := range pkg.Syntax {
				a.processFile(ctx.Block().WithFile(file), file, pkg)
			}

			return true
		})
	}
}

func (a *Analyzer) processFile(ctx *Context, file *ast.File, pkg *packages.Package) {
	comment := ParseComment(file.Doc)
	if comment.Ignore() {
		return
	}

	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncDecl:
			a.funDecl(ctx.Block(), node, file, pkg)
			return false
		case *ast.BlockStmt:
			a.blockStmt(ctx.Block(), node, file, pkg)
			return false
		}

		a.analyze(ctx, node)
		return true
	})
}

func (a *Analyzer) funDecl(ctx *Context, node *ast.FuncDecl, file *ast.File, pkg *packages.Package) {
	comment := ParseComment(node.Doc)
	if comment.Ignore() {
		return
	}

	ast.Inspect(node, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.BlockStmt:
			a.blockStmt(ctx.Block(), node, file, pkg)
			return false
		}

		a.analyze(ctx, node)
		return true
	})
}

func (a *Analyzer) loadDefinitionsFromPkg(pkg *packages.Package, moduleDir string) {
	InspectPackage(pkg, func(pkg *packages.Package) bool {
		if pkg.Module == nil { // Go 内置包
			ignore := true
			for _, depend := range a.depends {
				if strings.HasPrefix(pkg.PkgPath, depend) {
					ignore = false
					break
				}
			}
			if ignore {
				return false
			}
		} else {
			if pkg.Module.Dir != moduleDir && !lo.Contains(a.depends, pkg.Module.Path) {
				return false
			}
		}

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(node ast.Node) bool {
				switch node := node.(type) {
				case *ast.FuncDecl:
					a.definitions.Set(NewFuncDefinition(pkg, file, node))
					return false
				case *ast.TypeSpec:
					a.definitions.Set(NewTypeDefinition(pkg, file, node))
					return false
				}
				return true
			})
		}
		return true
	})
}

func (a *Analyzer) blockStmt(ctx *Context, node *ast.BlockStmt, file *ast.File, pkg *packages.Package) {
	comment := ParseComment(a.context().WithPackage(pkg).WithFile(file).GetHeadingCommentOf(node.Lbrace))
	if comment.Ignore() {
		return
	}

	a.analyze(ctx, node)

	for _, node := range node.List {
		ast.Inspect(node, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.BlockStmt:
				a.blockStmt(ctx.Block(), node, file, pkg)
				return false
			}

			a.analyze(ctx, node)
			return true
		})
	}
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
	dir, fileName := a.lookupGoModFile(pkgPath)
	if fileName == "" {
		panic("go.mod not found in " + pkgPath)
	}

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
		Dir:       dir,
		GoMod:     fileName,
		GoVersion: mod.Go.Version,
	}
}

func (a *Analyzer) lookupGoModFile(pkgPath string) (string, string) {
	for {
		fileName := filepath.Join(pkgPath, "go.mod")
		_, err := os.Stat(fileName)
		if err == nil {
			return strings.TrimSuffix(pkgPath, string(filepath.Separator)), fileName
		}
		var suffix string
		pkgPath, suffix = filepath.Split(pkgPath)
		if suffix == "" {
			break
		}
	}

	return "", ""
}

func (a *Analyzer) context() *Context {
	return newContext(a, a.globalEnv)
}

func (a *Analyzer) AddRoutes(items ...*API) {
	a.routes.add(items...)

	for _, item := range items {
		path := a.doc.Paths[item.FullPath]
		if path == nil {
			path = &spec.PathItem{}
		}
		item.applyToPathItem(path)
		a.doc.Paths[item.FullPath] = path
	}
}
