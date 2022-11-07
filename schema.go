package analyzer

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"github.com/gotomicro/ego-gen-api/tag"
	"github.com/samber/lo"

	"github.com/go-openapi/spec"
)

const (
	mimeTypeJson           = "application/json"
	mimeTypeXml            = "text/xml"
	mimeTypeFormData       = "multipart/form-data"
	mimeTypeFormUrlencoded = "application/x-www-form-urlencoded"
)

type SchemaBuilder struct {
	ctx         *Context
	contentType string
	stack       Stack[string]
}

func NewSchemaBuilder(ctx *Context, contentType string) *SchemaBuilder {
	return &SchemaBuilder{ctx: ctx, contentType: contentType}
}

func newSchemaBuilderWithStack(ctx *Context, contentType string, stack Stack[string]) *SchemaBuilder {
	return &SchemaBuilder{ctx: ctx, contentType: contentType, stack: stack}
}

func (s *SchemaBuilder) GetSchemaByExpr(expr ast.Expr, contentType string) *spec.Schema {
	t := s.ctx.Package().TypesInfo.TypeOf(expr)
	if t, ok := t.(*types.Basic); ok {
		return s.basicType(t.Name())
	}

	def := s.ctx.ParseType(t)
	typeDef, ok := def.(*TypeDefinition)
	if !ok {
		return nil
	}
	if lo.Contains(s.stack, typeDef.Key()) {
		return spec.RefSchema(typeDef.RefKey())
	}
	_, ok = s.ctx.Doc().Definitions[typeDef.ModelKey()]
	if ok {
		return spec.RefSchema(typeDef.RefKey())
	}

	s.stack.Push(typeDef.Key())
	defer s.stack.Pop()

	payloadSchema := newSchemaBuilderWithStack(s.ctx.WithPackage(typeDef.pkg).WithFile(typeDef.file), contentType, append(s.stack, typeDef.Key())).
		FromTypeSpec(typeDef.Spec)
	payloadSchema.ID = strings.ReplaceAll(typeDef.Key(), "/", "_")
	s.ctx.Doc().Definitions[typeDef.ModelKey()] = *payloadSchema

	return spec.RefSchema(typeDef.RefKey())
}

func (s *SchemaBuilder) FromTypeSpec(t *ast.TypeSpec) *spec.Schema {
	schema := s.parseExpr(t.Type)
	schema.Title = t.Name.Name
	schema.Description = NormalizeComment(t.Comment.Text(), t.Name.Name)
	return schema
}

func (s *SchemaBuilder) parseExpr(expr ast.Expr) (schema *spec.Schema) {
	switch expr := expr.(type) {
	case *ast.StructType:
		return s.parseStruct(expr)

	case *ast.StarExpr:
		return s.parseExpr(expr.X)

	case *ast.Ident:
		return s.parseIdent(expr)

	case *ast.SelectorExpr:
		return s.parseSelectorExpr(expr)

	case *ast.MapType:
		return spec.MapProperty(s.parseExpr(expr.Value))

	case *ast.ArrayType:
		return spec.ArrayProperty(s.parseExpr(expr.Elt))

	case *ast.SliceExpr:
		return spec.ArrayProperty(s.parseExpr(expr.X))
	}

	// TODO
	return nil
}

func (s *SchemaBuilder) parseStruct(expr *ast.StructType) *spec.Schema {
	schema := &spec.Schema{}
	schema.WithProperties(make(spec.SchemaProperties))

	var contentType = s.contentType
	if s.contentType == "" {
		contentType = "application/json" // fallback to json
	}

	for _, field := range expr.Fields.List {
		comment := ParseComment(field.Doc)
		for _, name := range field.Names {
			fieldSchema := s.parseExpr(field.Type)
			if fieldSchema == nil {
				fmt.Printf("unknown field type %s at %s\n", name.Name, s.ctx.LineColumn(field.Type.Pos()))
				continue
			}
			propName := s.getPropName(name.Name, field, contentType)
			if comment != nil {
				comment.transformIntoSchema(fieldSchema)
				if comment.Required() {
					schema.Required = append(schema.Required, propName)
				}
			}
			schema.Properties[propName] = *fieldSchema
		}
	}

	return schema
}

func (s *SchemaBuilder) parseIdent(expr *ast.Ident) *spec.Schema {
	schema := s.basicType(expr.Name)
	if schema != nil {
		return schema
	}

	return s.GetSchemaByExpr(expr, s.contentType)
}

func (s *SchemaBuilder) parseSelectorExpr(expr *ast.SelectorExpr) *spec.Schema {
	return s.parseExpr(expr.Sel)
}

func (s *SchemaBuilder) getPropName(fieldName string, field *ast.Field, contentType string) (propName string) {
	if field.Tag == nil {
		return fieldName
	}

	tags := tag.Parse(field.Tag.Value)
	var tagValue string
	switch contentType {
	case mimeTypeJson:
		tagValue = tags["json"]
	case mimeTypeXml:
		tagValue = tags["xml"]
	case mimeTypeFormData, mimeTypeFormUrlencoded:
		tagValue = tags["form"]
	}
	if tagValue == "" {
		return fieldName
	}

	propName, _, _ = strings.Cut(tagValue, ",")
	return
}

func (s *SchemaBuilder) basicType(name string) *spec.Schema {
	switch name {
	case "uint", "int", "uint8", "int8", "uint16", "int16",
		"uint32", "int32", "uint64", "int64",
		"byte", "rune":
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"integer"}}}
	case "float32", "float64":
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"number"}}}
	case "bool":
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"boolean"}}}
	case "string":
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"string"}}}
	}

	return nil
}