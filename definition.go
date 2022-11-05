package analyzer

import (
	"go/ast"
	"strings"

	"github.com/go-openapi/spec"
	"golang.org/x/tools/go/packages"
)

type Definition interface {
	Pkg() *packages.Package
	File() *ast.File
	Key() string

	definition()
}

var _ Definition = &FuncDefinition{}

type FuncDefinition struct {
	pkg  *packages.Package
	file *ast.File
	decl *ast.FuncDecl
}

func NewFuncDefinition(pkg *packages.Package, file *ast.File, decl *ast.FuncDecl) *FuncDefinition {
	return &FuncDefinition{pkg: pkg, file: file, decl: decl}
}

func (f *FuncDefinition) Key() string {
	return f.pkg.PkgPath + "." + f.decl.Name.Name
}

func (f *FuncDefinition) Pkg() *packages.Package {
	return f.pkg
}

func (f *FuncDefinition) File() *ast.File {
	return f.file
}

func (f *FuncDefinition) definition() {}

var _ Definition = &TypeDefinition{}

type TypeDefinition struct {
	Spec *ast.TypeSpec

	pkg  *packages.Package
	file *ast.File
}

func NewTypeDefinition(pkg *packages.Package, file *ast.File, spec *ast.TypeSpec) *TypeDefinition {
	return &TypeDefinition{pkg: pkg, file: file, Spec: spec}
}

func (t *TypeDefinition) definition() {}

func (t *TypeDefinition) Pkg() *packages.Package {
	return t.pkg
}

func (t *TypeDefinition) File() *ast.File {
	return t.file
}

func (t *TypeDefinition) Key() string {
	return t.pkg.PkgPath + "." + t.Spec.Name.Name
}

func (t *TypeDefinition) ToSwaggerSchema(ctx *Context, contentType string) *spec.Schema {
	return NewSchemaBuilder(ctx, contentType).FromTypeSpec(t.Spec)
}

func (t *TypeDefinition) ModelKey() string {
	return strings.ReplaceAll(t.pkg.PkgPath, "/", "_") + "." + t.Spec.Name.Name
}

func (t *TypeDefinition) RefKey() string {
	return "#/definitions/" + t.ModelKey()
}

type Definitions map[string]Definition

func (d *Definitions) Set(def Definition) {
	(*d)[def.Key()] = def
}

func (d *Definitions) Get(key string) Definition {
	return (*d)[key]
}
