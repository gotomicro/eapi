package eapi

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"strings"

	"github.com/gotomicro/eapi/spec"
	"github.com/gotomicro/eapi/tag"
	"github.com/iancoleman/strcase"
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
	typeArgs    []*spec.SchemaRef
	typeParams  []*spec.TypeParam
}

func NewSchemaBuilder(ctx *Context, contentType string) *SchemaBuilder {
	return &SchemaBuilder{
		ctx:         ctx,
		contentType: contentType,
	}
}

func newSchemaBuilderWithStack(ctx *Context, contentType string, stack Stack[string]) *SchemaBuilder {
	return &SchemaBuilder{ctx: ctx, contentType: contentType, stack: stack}
}

func (s *SchemaBuilder) clone() *SchemaBuilder {
	ret := *s
	return &ret
}

func (s *SchemaBuilder) parseTypeDef(def *TypeDefinition) *spec.SchemaRef {
	schemaRef := s.parseTypeSpec(def.Spec)
	if schemaRef == nil {
		return nil
	}
	schemaRef.Key = def.ModelKey(s.typeArgs...)

	if len(def.Enums) > 0 {
		schema := spec.Unref(s.ctx.Doc(), schemaRef)
		ext := spec.NewExtendedEnumType(def.Enums...)
		schema.ExtendedTypeInfo = ext
		for _, item := range def.Enums {
			schema.Enum = append(schema.Enum, item.Value)
		}
	}

	return schemaRef
}

func (s *SchemaBuilder) parseTypeSpec(t *ast.TypeSpec) *spec.SchemaRef {
	var typeParams []*spec.TypeParam
	if t.TypeParams != nil {
		for i, field := range t.TypeParams.List {
			for j, name := range field.Names {
				typeParams = append(typeParams, &spec.TypeParam{
					Index:      i + j,
					Name:       name.Name,
					Constraint: field.Type.(*ast.Ident).Name,
				})
			}
		}
	}

	schema := s.setTypeParams(typeParams).ParseExpr(t.Type)
	if schema == nil {
		return nil
	}
	if t.TypeParams != nil {
		schema.ExtendedTypeInfo.TypeParams = typeParams
	}

	comment := s.ctx.ParseComment(s.ctx.GetHeadingCommentOf(t.Type.Pos()))
	comment.ApplyToSchema(schema)
	if schema.Ref == "" {
		schema.Title = strcase.ToCamel(s.ctx.Package().Name + t.Name.Name)
	}
	return schema
}

func (s *SchemaBuilder) setTypeParams(params []*spec.TypeParam) *SchemaBuilder {
	s.typeParams = params
	return s
}

func (s *SchemaBuilder) setTypeArgs(args ...*spec.SchemaRef) *SchemaBuilder {
	res := *s
	res.typeArgs = args
	return &res
}

func (s *SchemaBuilder) ParseExpr(expr ast.Expr) (schema *spec.SchemaRef) {
	switch expr := expr.(type) {
	case *ast.BasicLit:
		return s.parseBasicLit(expr)

	case *ast.StructType:
		return s.parseStruct(expr)

	case *ast.StarExpr:
		return s.ParseExpr(expr.X)

	case *ast.Ident:
		return s.parseIdent(expr)

	case *ast.SelectorExpr:
		return s.ParseExpr(expr.Sel)

	case *ast.MapType:
		value := s.ParseExpr(expr.Value)
		return spec.NewObjectSchema().
			WithExtendedType(spec.NewMapExtendedType(
				s.ParseExpr(expr.Key),
				value,
			)).
			WithAdditionalProperties(value)

	case *ast.ArrayType:
		return spec.NewArraySchema(s.ParseExpr(expr.Elt)).NewRef()

	case *ast.SliceExpr:
		return spec.NewArraySchema(s.ParseExpr(expr.X)).NewRef()

	case *ast.UnaryExpr:
		return s.ParseExpr(expr.X)

	case *ast.CompositeLit:
		return s.ParseExpr(expr.Type)

	case *ast.InterfaceType:
		return spec.NewObjectSchema().WithDescription("Any Type").WithExtendedType(spec.NewAnyExtendedType())

	case *ast.CallExpr:
		return s.parseCallExpr(expr)

	case *ast.IndexExpr:
		return s.parseIndexExpr(expr)

	case *ast.IndexListExpr:
		return s.parseIndexListExpr(expr)

	}

	fmt.Fprintf(os.Stderr, "unknown type at %s\n", s.ctx.LineColumn(expr.Pos()))
	return spec.NewObjectSchema().WithExtendedType(spec.NewUnknownExtType()).NewRef()
}

func (s *SchemaBuilder) parseBasicLit(expr *ast.BasicLit) *spec.SchemaRef {
	switch expr.Kind {
	case token.INT:
		return s.basicType("int")
	case token.FLOAT:
		return s.basicType("float64")
	case token.IMAG:
		return s.basicType("float64")
	case token.CHAR:
		return s.basicType("string")
	case token.STRING:
		return s.basicType("string")
	}

	fmt.Fprintf(os.Stderr, "unknown type at %s\n", s.ctx.LineColumn(expr.Pos()))
	return spec.NewObjectSchema().WithExtendedType(spec.NewUnknownExtType()).NewRef()
}

func (s *SchemaBuilder) parseStruct(expr *ast.StructType) *spec.SchemaRef {
	schema := spec.NewObjectSchema()
	schema.Properties = make(spec.Schemas)

	var contentType = s.contentType
	if s.contentType == "" {
		contentType = "application/json" // fallback to json
	}

	for _, field := range expr.Fields.List {
		comment := s.parseCommentOfField(field)
		if comment.Ignore() {
			continue // ignored field
		}

		if len(field.Names) == 0 { // type composition
			fieldSchema := s.ParseExpr(field.Type)
			if fieldSchema != nil {
				// merge properties
				fieldSchema = spec.Unref(s.ctx.Doc(), fieldSchema)
				if fieldSchema != nil {
					for name, value := range fieldSchema.Properties {
						schema.Properties[name] = value
					}
				}
			}
		}

		for _, name := range field.Names {
			if !name.IsExported() {
				continue
			}
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

	return schema
}

func (s *SchemaBuilder) parseIdent(expr *ast.Ident) *spec.SchemaRef {
	t := s.ctx.Package().TypesInfo.TypeOf(expr)
	return s.parseType(t)
}

var commonTypes = map[string]*spec.Schema{
	"time.Time": spec.NewSchema().WithType("string").WithFormat("datetime"),
	"encoding/json.RawMessage": spec.NewSchema().
		WithType("object").
		WithDescription("Any Json Type").
		WithExtendedType(spec.NewAnyExtendedType()),
	"json.RawMessage": spec.NewSchema().
		WithType("object").
		WithDescription("Any Json Type").
		WithExtendedType(spec.NewAnyExtendedType()),
	"database/sql.NullTime":    spec.NewDateTimeSchema(),
	"database/sql.NullString":  spec.NewStringSchema(),
	"database/sql.NullInt64":   spec.NewInt64Schema(),
	"database/sql.NullInt32":   spec.NewInt32Schema(),
	"database/sql.NullInt":     spec.NewIntegerSchema(),
	"database/sql.NullInt16":   spec.NewIntegerSchema(),
	"database/sql.NullFloat64": spec.NewFloat64Schema(),
	"database/sql.NullBool":    spec.NewBoolSchema(),
	"database/sql.NullByte":    spec.NewStringSchema(),
}

func (s *SchemaBuilder) commonUsedType(t types.Type) *spec.SchemaRef {
	switch t := t.(type) {
	case *types.Named:
		typeName := t.Obj().Pkg().Path() + "." + t.Obj().Name()
		commonType, ok := commonTypes[typeName]
		if !ok {
			return nil
		}
		return commonType.Clone()

	case *types.Pointer:
		return s.commonUsedType(t.Elem())
	}

	return nil
}

func (s *SchemaBuilder) parseSelectorExpr(expr *ast.SelectorExpr) *spec.SchemaRef {
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

func (s *SchemaBuilder) basicType(name string) *spec.SchemaRef {
	switch name {
	case "uint", "int", "uint8", "int8", "uint16", "int16",
		"uint32", "int32", "uint64", "int64":
		return spec.NewIntegerSchema()
	case "byte", "rune":
		return spec.NewBytesSchema()
	case "float32", "float64":
		return spec.NewFloat64Schema()
	case "bool":
		return spec.NewBoolSchema()
	case "string":
		return spec.NewStringSchema()
	}

	return nil
}

func (s *SchemaBuilder) inParsingStack(key string) bool {
	return lo.Contains(s.stack, key)
}

func (s *SchemaBuilder) parseType(t types.Type) *spec.SchemaRef {
	var typeArgs = s.typeArgs

	switch t := t.(type) {
	case *types.Basic:
		return s.basicType(t.Name())
	case *types.Interface:
		return spec.NewObjectSchema().
			WithDescription("Any Type").
			WithExtendedType(spec.NewAnyExtendedType()).NewRef()
	case *types.TypeParam:
		return spec.NewTypeParamSchema(s.typeParams[t.Index()]).NewRef()
	case *types.Slice:
		return spec.NewArraySchema(s.parseType(t.Elem())).NewRef()
	case *types.Array:
		return spec.NewArraySchema(s.parseType(t.Elem())).NewRef()
	case *types.Pointer:
		return s.parseType(t.Elem())
	case *types.Map:
		valueType := s.parseType(t.Elem())
		keyType := s.parseType(t.Key())
		return spec.NewObjectSchema().
			WithAdditionalProperties(valueType).
			WithExtendedType(spec.NewMapExtendedType(keyType, valueType)).
			NewRef()
	case *types.Named:
		// parse type arguments
		args := t.TypeArgs()
		if args.Len() > 0 {
			typeArgs = make([]*spec.SchemaRef, 0)
			for i := 0; i < args.Len(); i++ {
				typeArgs = append(typeArgs, s.parseType(args.At(i)))
			}
		}
	default:
	}

	// 检查是否是常用类型
	schema := s.commonUsedType(t)
	if schema != nil {
		return schema
	}

	def := s.ctx.ParseType(t)
	typeDef, ok := def.(*TypeDefinition)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown type %s\n", t.String())
		return spec.NewSchema().WithExtendedType(spec.NewUnknownExtType()).NewRef()
	}

	modelKey := typeDef.ModelKey()
	refKey := typeDef.RefKey()
	if s.inParsingStack(modelKey) {
		return spec.RefSchema(refKey)
	}

	schema, schemaExists := s.ctx.Doc().Components.Schemas[modelKey]
	if !schemaExists {
		s.stack.Push(modelKey)
		defer s.stack.Pop()

		schema = newSchemaBuilderWithStack(s.ctx.WithPackage(typeDef.pkg).WithFile(typeDef.file), s.contentType, s.stack).
			setTypeArgs().
			parseTypeDef(typeDef)
		s.ctx.Doc().Components.Schemas[typeDef.ModelKey()] = schema
	}

	schemaRef := spec.RefSchema(refKey)
	if len(typeArgs) > 0 {
		return spec.NewSchema().WithExtendedType(spec.NewSpecificExtendType(schemaRef, typeArgs...)).NewRef()
	}
	return schemaRef
}

func (s *SchemaBuilder) parseCommentOfField(field *ast.Field) *Comment {
	// heading comment
	if field.Doc != nil && len(field.Doc.List) > 0 {
		return s.ctx.ParseComment(field.Doc)
	}

	// parse trailing comment
	commentGroup := s.ctx.GetTrailingCommentOf(field.Pos())
	return s.ctx.ParseComment(commentGroup)
}

func (s *SchemaBuilder) parseCallExpr(expr *ast.CallExpr) *spec.SchemaRef {
	t := s.ctx.Package().TypesInfo.TypeOf(expr)
	return s.parseType(t)
}

func (s *SchemaBuilder) parseIndexExpr(expr *ast.IndexExpr) *spec.SchemaRef {
	var argType *spec.SchemaRef
	if s.isTypeParam(expr.Index) {
		t := s.ctx.Package().TypesInfo.TypeOf(expr.Index).(*types.TypeParam)
		argType = spec.NewTypeParamSchema(s.typeParams[t.Index()]).NewRef()
	} else {
		argType = s.ParseExpr(expr.Index)
	}
	genericType := s.clone().setTypeArgs().setTypeParams(nil).ParseExpr(expr.X)
	if genericType == nil {
		return nil
	}

	return spec.NewSchema().WithExtendedType(spec.NewSpecificExtendType(genericType, argType)).NewRef()
}

func (s *SchemaBuilder) parseIndexListExpr(expr *ast.IndexListExpr) *spec.SchemaRef {
	var typeArgs []*spec.SchemaRef
	for _, param := range expr.Indices {
		var typeArg *spec.SchemaRef
		if s.isTypeParam(param) {
			t := s.ctx.Package().TypesInfo.TypeOf(param).(*types.TypeParam)
			typeArg = spec.NewTypeParamSchema(s.typeParams[t.Index()]).NewRef()
		} else {
			typeArg = s.ParseExpr(param)
		}
		typeArgs = append(typeArgs, typeArg)
	}
	genericType := s.setTypeArgs().setTypeParams(nil).ParseExpr(expr.X)
	if genericType == nil {
		return nil
	}

	return spec.NewSchema().WithExtendedType(spec.NewSpecificExtendType(genericType, typeArgs...)).NewRef()
}

func (s *SchemaBuilder) getTypeKey(expr ast.Expr) string {
	t := s.ctx.Package().TypesInfo.TypeOf(expr)
	switch t := t.(type) {
	case *types.Basic:
		return t.Name()
	default:
		def := s.ctx.ParseType(t)
		if def == nil {
			fmt.Printf("unknown type at %s\n", s.ctx.LineColumn(expr.Pos()))
			return ""
		}
		return def.(*TypeDefinition).ModelKey()
	}
}

// 判断表达式是否是泛型类型形参
func (s *SchemaBuilder) isTypeParam(index ast.Expr) bool {
	t := s.ctx.Package().TypesInfo.TypeOf(index)
	_, ok := t.(*types.TypeParam)
	return ok
}
