package analyzer

import (
	"go/ast"
	"go/types"

	"github.com/go-openapi/spec"
	"github.com/gotomicro/ego-gen-api/tag"
)

type ParamParser struct {
	ctx         *Context
	contentType string
}

func NewParamParser(ctx *Context, contentType string) *ParamParser {
	return &ParamParser{ctx: ctx, contentType: contentType}
}

// Parse 根据 ast.Expr 解析出 []*spec.Parameter
func (p *ParamParser) Parse(expr ast.Expr) (params []*spec.Parameter) {
	switch expr := expr.(type) {
	case *ast.Ident:
		return p.parseIdent(expr)
	case *ast.StarExpr:
		return p.Parse(expr.X)
	case *ast.UnaryExpr:
		return p.Parse(expr.X)
	case *ast.SelectorExpr:
		return p.Parse(expr.Sel)
	}
	return
}

func (p *ParamParser) parseIdent(expr *ast.Ident) (params []*spec.Parameter) {
	t := p.ctx.Package().TypesInfo.TypeOf(expr)
	if t == nil {
		return
	}

	return p.parseType(t)
}

func (p *ParamParser) parseType(t types.Type) (params []*spec.Parameter) {
	def := p.ctx.ParseType(t)
	typeDef, ok := def.(*TypeDefinition)
	if !ok {
		return nil
	}
	structType, ok := typeDef.Spec.Type.(*ast.StructType)
	if !ok {
		return
	}

	params = NewParamParser(
		p.ctx.WithPackage(typeDef.Pkg()).WithFile(typeDef.File()),
		p.contentType,
	).parseStructType(structType)
	return
}

func (p *ParamParser) parseStructType(structType *ast.StructType) (params []*spec.Parameter) {
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			param := p.parseField(name, field)
			params = append(params, param)
		}
	}
	return
}

func (p *ParamParser) parseField(name *ast.Ident, field *ast.Field) (param *spec.Parameter) {
	param = &spec.Parameter{}
	param.Type, param.Format = p.typeOf(field.Type)

	tagValues := tag.Parse(field.Tag.Value)
	formName, ok := tagValues["form"]
	if !ok {
		formName = name.Name
	}
	param.Name = formName

	// parse comments
	comments := ParseComment(field.Doc)
	if comments != nil {
		param.Required = comments.Required()
		param.Nullable = comments.Nullable()
		param.Description = comments.Text
	}

	return
}

func (p *ParamParser) typeOf(expr ast.Expr) (string, string) {
	switch t := expr.(type) {
	case *ast.Ident:
		return p.typeOfIdent(t)

	case *ast.SelectorExpr:
		return p.typeOf(t.Sel)
	}

	// fallback
	return "string", ""
}

func (p *ParamParser) typeOfIdent(ident *ast.Ident) (string, string) {
	paramType := p.basicType(ident.Name)
	if paramType != "" {
		return paramType, ""
	}

	t := p.ctx.Package().TypesInfo.TypeOf(ident)
	if t == nil {
		// unknown
		return "", ""
	}

	return p.parseTypeOfType(t)
}

// Swagger Parameter types:
// Name		|	type	| 	format		|	Comments
// integer	|	integer |	int32		|	signed 32 bits
// long		|	integer |	int64		|	signed 64 bits
// float	|	number 	|	float		|
// double	|	number 	|	double		|
// string	|	string 	|				|
// byte		|	string 	|	byte		|	base64 encoded characters
// binary	|	string 	|	binary		|	any sequence of octets
// boolean	|	boolean |				|
// date		|	string 	|	date		|	As defined by full-date - RFC3339
// dateTime	|	string 	|	date-time	|	As defined by date-time - RFC3339
// password	|	string 	|	password	|	Used to hint UIs the input needs to be obscured.
func (p *ParamParser) basicType(name string) string {
	switch name {
	case "uint", "int", "uint8", "int8", "uint16", "int16",
		"uint32", "int32", "uint64", "int64",
		"byte", "rune":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "string":
		return "string"
	}

	return ""
}

// type:
//		The value MUST be one of "string", "number", "integer", "boolean", "array" or "file".
//		If type is "file", the consumes MUST be either "multipart/form-data", " application/x-www-form-urlencoded" or
//		both and the parameter MUST be in "formData".
func (p *ParamParser) param(name, paramType string, repeated bool) {
	param := &spec.Parameter{}
	param.Type = ""
}

func (p *ParamParser) parseTypeOfType(t types.Type) (string, string) {
	switch t := t.(type) {
	case *types.Named:
		p.parseTypeOfType(t.Underlying())
	case *types.Basic:
		return p.parseTypeOfBasicType(t)
	}

	return "", ""
}

func (p *ParamParser) parseTypeOfBasicType(t *types.Basic) (string, string) {
	switch t.Kind() {
	case types.Bool:
		return "boolean", ""
	case types.Int,
		types.Int8,
		types.Int16,
		types.Int32,
		types.Int64,
		types.Uint,
		types.Uint8,
		types.Uint16,
		types.Uint32,
		types.Uint64,
		types.Uintptr:
		return "integer", ""
	case types.Float32,
		types.Float64,
		types.Complex64,
		types.Complex128:
		return "number", ""
	case types.String:
		return "string", ""
	}

	// unknown
	return "", ""
}