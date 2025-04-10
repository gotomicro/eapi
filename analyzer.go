package eapi

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotomicro/eapi/spec"
	"github.com/knadh/koanf"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

type Analyzer struct {
	routes      APIs
	globalEnv   *Environment
	plugins     []Plugin
	definitions Definitions
	depends     []string
	k           *koanf.Koanf

	doc      *spec.T
	packages []*packages.Package
}

func NewAnalyzer(k *koanf.Koanf) *Analyzer {
	a := &Analyzer{
		routes:      make(APIs, 0),
		globalEnv:   NewEnvironment(nil),
		plugins:     make([]Plugin, 0),
		definitions: make(Definitions),
		k:           k,
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
		err := plugin.Mount(a.k)
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

func (a *Analyzer) Process(packagePath string) *Analyzer {
	if len(a.plugins) <= 0 {
		panic("must register plugin before processing")
	}

	packagePath, err := filepath.Abs(packagePath)
	if err != nil {
		panic("invalid package path: " + err.Error())
	}

	var visited = make(map[string]struct{})
	pkgList := a.load(packagePath)
	for _, pkg := range pkgList {
		a.definitions = make(Definitions)
		for _, p := range pkg {
			a.loadDefinitionsFromPkg(p, p.Module.Dir)
		}

		for _, pkg := range pkg {
			moduleDir := pkg.Module.Dir
			InspectPackage(pkg, func(pkg *packages.Package) bool {
				if _, ok := visited[pkg.PkgPath]; ok {
					return false
				}
				visited[pkg.PkgPath] = struct{}{}
				if pkg.Module == nil || pkg.Module.Dir != moduleDir {
					return false
				}
				if DEBUG {
					fmt.Printf("inspect %s\n", pkg.PkgPath)
				}

				ctx := a.context().Block().WithPackage(pkg)
				for _, file := range pkg.Syntax {
					a.processFile(ctx.Block().WithFile(file), file, pkg)
				}

				return true
			})
		}
	}

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

const entryPackageName = "command-line-arguments"

func (a *Analyzer) load(pkgPath string) [][]*packages.Package {
	absPath, err := filepath.Abs(pkgPath)
	if err != nil {
		panic("invalid package path: " + pkgPath)
	}

	var pkgList []*build.Package
	filepath.Walk(absPath, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		pkg, err := build.Default.ImportDir(path, build.ImportComment)
		if err != nil {
			var noGoErr = &build.NoGoError{}
			if errors.As(err, &noGoErr) {
				return nil
			}
			panic("import directory failed: " + err.Error())
		}
		pkgList = append(pkgList, pkg)
		return filepath.SkipDir
	})

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
	var res [][]*packages.Package
	for _, pkg := range pkgList {
		var files []string
		for _, filename := range append(pkg.GoFiles, pkg.CgoFiles...) {
			files = append(files, filepath.Join(pkg.Dir, filename))
		}
		packs, err := packages.Load(config, files...)
		if err != nil {
			panic("load packages failed: " + err.Error())
		}

		// 前面的 packages.Load() 方法不能解析出以第一层的 Module
		// 所以这里手动解析 go.mod
		for _, p := range packs {
			if p.Module != nil {
				continue
			}

			module := a.parseGoModule(pkgPath)
			if module == nil {
				panic("failed to parse go.mod file in " + pkgPath)
			}
			p.Module = module
			p.PkgPath = entryPackageName
			p.ID = module.Path
		}
		res = append(res, packs)
	}

	return res
}

func (a *Analyzer) processFile(ctx *Context, file *ast.File, pkg *packages.Package) {
	comment := ctx.ParseComment(file.Doc)
	if comment.Ignore() {
		return
	}
	ctx.commentStack.comment = comment

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
	comment := ctx.ParseComment(node.Doc)
	if comment.Ignore() {
		return
	}
	ctx.commentStack.comment = comment

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
	var visited = make(map[string]struct{})
	InspectPackage(pkg, func(pkg *packages.Package) bool {
		if _, ok := visited[pkg.PkgPath]; ok {
			return false
		}
		visited[pkg.PkgPath] = struct{}{}

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
				case *ast.GenDecl:
					if node.Tok == token.CONST {
						a.loadEnumDefinition(pkg, file, node)
						return false
					}
					return true
				}
				return true
			})
		}
		return true
	})
}

type A int

const (
	A1 A = iota + 1
	A2
	A3
)

func (a *Analyzer) loadEnumDefinition(pkg *packages.Package, file *ast.File, node *ast.GenDecl) {
	for _, item := range node.Specs {
		valueSpec, ok := item.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range valueSpec.Names {
			c := pkg.TypesInfo.ObjectOf(name).(*types.Const)
			t, ok := c.Type().(*types.Named)
			if !ok {
				continue
			}
			basicType, ok := t.Underlying().(*types.Basic)
			if !ok {
				continue
			}
			pkgPath := t.Obj().Pkg().Path()
			if pkgPath != pkg.PkgPath {
				continue
			}
			def := a.definitions.Get(t.Obj().Pkg().Path() + "." + t.Obj().Name())
			if def == nil {
				continue
			}
			typeDef := def.(*TypeDefinition)
			value := ConvertStrToBasicType(c.Val().ExactString(), basicType)
			enumItem := spec.NewExtendEnumItem(name.Name, value, strings.TrimSpace(valueSpec.Doc.Text()))
			typeDef.Enums = append(typeDef.Enums, enumItem)
		}
	}
}

func (a *Analyzer) blockStmt(ctx *Context, node *ast.BlockStmt, file *ast.File, pkg *packages.Package) {
	comment := ctx.ParseComment(a.context().WithPackage(pkg).WithFile(file).GetHeadingCommentOf(node.Lbrace))
	if comment.Ignore() {
		return
	}
	ctx.commentStack.comment = comment

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

func (a *Analyzer) parseGoModule(pkgPath string) *packages.Module {
	dir, fileName := a.lookupGoModFile(pkgPath)
	if fileName == "" {
		panic("go.mod not found in " + pkgPath)
	}

	content, err := os.ReadFile(fileName)
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
