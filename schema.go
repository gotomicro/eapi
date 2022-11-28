package analyzer

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gotomicro/ego-gen-api/spec"
	"github.com/gotomicro/ego-gen-api/tag"
	"github.com/samber/lo"
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

func (s *SchemaBuilder) FromTypeSpec(t *ast.TypeSpec) *openapi3.SchemaRef {
	schema := s.ParseExpr(t.Type)
	if schema == nil {
		return nil
	}
	schema.Value.Title = t.Name.Name
	schema.Value.Description = NormalizeComment(t.Comment.Text(), t.Name.Name)
	return schema
}

func (s *SchemaBuilder) ParseExpr(expr ast.Expr) (schema *openapi3.SchemaRef) {
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

	case *ast.InterfaceType:
		return openapi3.NewSchemaRef("", openapi3.NewSchema())

	case *ast.CallExpr:
		return s.parseCallExpr(expr)
	}

	// TODO
	return nil
}

func (s *SchemaBuilder) parseStruct(expr *ast.StructType) *openapi3.SchemaRef {
	schema := openapi3.NewObjectSchema()
	schema.Properties = make(openapi3.Schemas)

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
				fieldSchema = spec.Unref(s.ctx.Doc(), fieldSchema)
				if fieldSchema.Value != nil {
					for name, value := range fieldSchema.Value.Properties {
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
				comment.ApplyToSchema(fieldSchema)
				if comment.Required() {
					schema.Required = append(schema.Required, propName)
				}
			}
			schema.Properties[propName] = fieldSchema
		}
	}

	return openapi3.NewSchemaRef("", schema)
}

func (s *SchemaBuilder) parseIdent(expr *ast.Ident) *openapi3.SchemaRef {
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

func (s *SchemaBuilder) commonUsedType(t types.Type) *openapi3.SchemaRef {
	switch t := t.(type) {
	case *types.Named:
		typeName := t.Obj().Pkg().Path() + "." + t.Obj().Name()
		commonType, ok := commonTypes[typeName]
		if !ok {
			return nil
		}
		schema := openapi3.NewSchema()
		schema.Type = commonType[0]
		schema.Format = commonType[1]
		return openapi3.NewSchemaRef("", schema)

	case *types.Pointer:
		return s.commonUsedType(t.Elem())
	}

	return nil
}

func (s *SchemaBuilder) parseSelectorExpr(expr *ast.SelectorExpr) *openapi3.SchemaRef {
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

func (s *SchemaBuilder) basicType(name string) *openapi3.SchemaRef {
	switch name {
	case "uint", "int", "uint8", "int8", "uint16", "int16",
		"uint32", "int32", "uint64", "int64":
		return openapi3.NewSchemaRef("", openapi3.NewIntegerSchema())
	case "byte", "rune":
		return openapi3.NewSchemaRef("", openapi3.NewBytesSchema())
	case "float32", "float64":
		return openapi3.NewSchemaRef("", openapi3.NewFloat64Schema())
	case "bool":
		return openapi3.NewSchemaRef("", openapi3.NewBoolSchema())
	case "string":
		return openapi3.NewSchemaRef("", openapi3.NewStringSchema())
	}

	return nil
}

func (s *SchemaBuilder) parseType(t types.Type) *openapi3.SchemaRef {
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

	_, ok = s.ctx.Doc().Components.Schemas[typeDef.ModelKey()]
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
	s.ctx.Doc().Components.Schemas[typeDef.ModelKey()] = payloadSchema

	return spec.RefSchema(typeDef.RefKey())
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

func (s *SchemaBuilder) parseCallExpr(expr *ast.CallExpr) *openapi3.SchemaRef {
	return s.parseFunReturnVal(expr.Fun)
}

func (s *SchemaBuilder) parseFunReturnVal(fun ast.Expr) *openapi3.SchemaRef {
	switch fun := fun.(type) {
	case *ast.SelectorExpr:
		return s.parseFunReturnVal(fun.Sel)
	case *ast.Ident:
		t := s.ctx.Package().TypesInfo.TypeOf(fun)
		switch t := t.(type) {
		case *types.Signature:
			callInfo := s.ctx.parseCallInfoByIdent(fun)
			if callInfo == nil {
				return nil
			}
			def := s.ctx.GetDefinition(callInfo.Type, callInfo.Method)
			if def == nil {
				fmt.Printf("unknown function %s.%s at %s\n", callInfo.Type, callInfo.Method, s.ctx.LineColumn(fun.Pos()))
				return nil
			}
			fnDef, ok := def.(*FuncDefinition)
			if !ok {
				return nil
			}
			if fnDef.Decl.Type.Results.NumFields() == 0 {
				return nil
			}
			ret := fnDef.Decl.Type.Results.List[0]
			return newSchemaBuilderWithStack(s.ctx.WithPackage(fnDef.pkg).WithFile(fnDef.file), s.contentType, append(s.stack, fnDef.Key())).
				ParseExpr(ret.Type)
		default:
			return s.parseType(t)
		}
	}

	return nil
}
