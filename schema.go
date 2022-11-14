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
	MimeTypeJson           = "application/json"
	MimeApplicationXml     = "application/xml"
	MimeTypeXml            = "text/xml"
	MimeTypeFormData       = "multipart/form-data"
	MimeTypeFormUrlencoded = "application/x-www-form-urlencoded"
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

func (s *SchemaBuilder) FromTypeSpec(t *ast.TypeSpec) *spec.Schema {
	schema := s.ParseExpr(t.Type)
	if schema == nil {
		return nil
	}
	schema.Title = t.Name.Name
	schema.Description = NormalizeComment(t.Comment.Text(), t.Name.Name)
	return schema
}

func (s *SchemaBuilder) ParseExpr(expr ast.Expr) (schema *spec.Schema) {
	switch expr := expr.(type) {
	case *ast.StructType:
		return s.parseStruct(expr)

	case *ast.StarExpr:
		return s.ParseExpr(expr.X)

	case *ast.Ident:
		return s.parseIdent(expr)

	case *ast.SelectorExpr:
		return s.ParseExpr(expr.Sel)

	case *ast.MapType:
		return spec.MapProperty(s.ParseExpr(expr.Value))

	case *ast.ArrayType:
		return spec.ArrayProperty(s.ParseExpr(expr.Elt))

	case *ast.SliceExpr:
		return spec.ArrayProperty(s.ParseExpr(expr.X))

	case *ast.UnaryExpr:
		return s.ParseExpr(expr.X)

	case *ast.CompositeLit:
		return s.ParseExpr(expr.Type)
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
		comment := s.parseCommentOfField(field)
		if comment != nil && comment.Ignore() {
			continue // ignored field
		}

		if len(field.Names) == 0 { // type composition
			fieldSchema := s.ParseExpr(field.Type)
			if fieldSchema != nil {
				// merge properties
				fieldSchema = s.unRef(fieldSchema)
				if fieldSchema != nil {
					for name, value := range fieldSchema.Properties {
						schema.Properties[name] = value
					}
				}
			}
		}

		for _, name := range field.Names {
			fieldSchema := s.ParseExpr(field.Type)
			if fieldSchema == nil {
				fmt.Printf("unknown field type %s at %s\n", name.Name, s.ctx.LineColumn(field.Type.Pos()))
				continue
			}
			propName := s.getPropName(name.Name, field, contentType)
			if propName == "-" { // ignore
				continue
			}

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
	t := s.ctx.Package().TypesInfo.TypeOf(expr)
	switch t := t.(type) {
	case *types.Basic:
		return s.basicType(t.Name())
	}

	// 检查是否是常用类型
	schema := s.commonUsedType(t)
	if schema != nil {
		return schema
	}

	return s.parseType(t)
}

var commonTypes = map[string][]string{
	"time.Time":                {"string", "datetime"},
	"encoding/json.RawMessage": {"string", "byte"},
}

func (s *SchemaBuilder) commonUsedType(t types.Type) *spec.Schema {
	switch t := t.(type) {
	case *types.Named:
		typeName := t.Obj().Pkg().Path() + "." + t.Obj().Name()
		commonType, ok := commonTypes[typeName]
		if !ok {
			return nil
		}
		schema := &spec.Schema{}
		schema.Type = append(schema.Type, commonType[0])
		schema.Format = commonType[1]
		return schema

	case *types.Pointer:
		return s.commonUsedType(t.Elem())
	}

	return nil
}

func (s *SchemaBuilder) parseSelectorExpr(expr *ast.SelectorExpr) *spec.Schema {
	return s.ParseExpr(expr.Sel)
}

func (s *SchemaBuilder) getPropName(fieldName string, field *ast.Field, contentType string) (propName string) {
	if field.Tag == nil {
		return fieldName
	}

	tags := tag.Parse(field.Tag.Value)
	var tagValue string
	switch contentType {
	case MimeTypeJson:
		tagValue = tags["json"]
	case MimeTypeXml, MimeApplicationXml:
		tagValue = tags["xml"]
	case MimeTypeFormData, MimeTypeFormUrlencoded:
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

func (s *SchemaBuilder) parseType(t types.Type) *spec.Schema {
	switch t := t.(type) {
	case *types.Slice:
		return spec.ArrayProperty(s.parseType(t.Elem()))
	case *types.Array:
		return spec.ArrayProperty(s.parseType(t.Elem()))
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

	payloadSchema := newSchemaBuilderWithStack(s.ctx.WithPackage(typeDef.pkg).WithFile(typeDef.file), s.contentType, append(s.stack, typeDef.Key())).
		FromTypeSpec(typeDef.Spec)
	if payloadSchema == nil {
		return nil
	}

	payloadSchema.ID = strings.ReplaceAll(typeDef.Key(), "/", "_")
	s.ctx.Doc().Definitions[typeDef.ModelKey()] = *payloadSchema

	return spec.RefSchema(typeDef.RefKey())
}

func (s *SchemaBuilder) unRef(schema *spec.Schema) *spec.Schema {
	ref := schema.Ref
	if ref.GetURL() == nil {
		return schema
	}

	tokens := ref.GetPointer().DecodedTokens()
	if len(tokens) != 2 {
		return nil
	}
	if tokens[0] != "definitions" {
		return nil
	}

	def, ok := s.ctx.Doc().Definitions[tokens[1]]
	if !ok {
		return nil
	}
	return &def
}

func (s *SchemaBuilder) parseCommentOfField(field *ast.Field) *Comment {
	// heading comment
	if field.Doc != nil && len(field.Doc.List) > 0 {
		return ParseComment(field.Doc)
	}

	// parse trailing comment
	commentGroup := s.ctx.GetTrailingCommentOf(field.Pos())
	return ParseComment(commentGroup)
}
