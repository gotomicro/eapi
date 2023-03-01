package common

import (
	"github.com/gotomicro/eapi/spec"
	"github.com/gotomicro/eapi/utils"
)

type DataSchemaTransformer struct {
	dataSchema *DataSchema
}

func NewDataSchemaTransformer(dataSchema *DataSchema) *DataSchemaTransformer {
	return &DataSchemaTransformer{dataSchema: dataSchema}
}

func (d *DataSchemaTransformer) TransformToGeneric() *spec.SchemaRef {
	var typeArgs = make([]*spec.TypeParam, 0)
	var transform func(dataType *DataSchema) *spec.SchemaRef
	transform = func(dataType *DataSchema) *spec.SchemaRef {
		if dataType == nil {
			return nil
		}
		switch dataType.Type {
		case DataTypeString:
			return d.basicSchemaType("string")
		case DataTypeNumber:
			return d.basicSchemaType("number")
		case DataTypeInteger:
			return d.basicSchemaType("integer")
		case DataTypeBoolean:
			return d.basicSchemaType("boolean")
		case DataTypeFile:
			return d.basicSchemaType("file")
		case DataTypeArray:
			return transform(dataType.Item)
		case DataTypeObject:
			schema := spec.NewObjectSchema()
			properties := make(spec.Schemas)
			utils.RangeMapInOrder(
				dataType.Properties,
				func(a, b string) bool { return a < b },
				func(name string, dataSchema *DataSchema) {
					if !dataSchema.Optional {
						schema.Required = append(schema.Required, name)
					}
					s := transform(dataSchema)
					if s != nil {
						properties[name] = s
					}
				},
			)
			schema.Properties = properties
			return schema
		default:
			typeArg := &spec.TypeParam{
				Index:      len(typeArgs),
				Name:       string(rune('A' + len(typeArgs))),
				Constraint: "any",
			}
			typeArgs = append(typeArgs, typeArg)
			res := spec.NewTypeParamSchema(typeArg).NewRef()
			return res
		}
	}
	res := transform(d.dataSchema)
	res.ExtendedTypeInfo.TypeParams = typeArgs
	return res
}

func (d *DataSchemaTransformer) TransformToSpecific(genericType *spec.SchemaRef, dataTypeFallback func(dataType *DataSchema) *spec.SchemaRef) *spec.SchemaRef {
	var typeArgs = make([]*spec.SchemaRef, 0)
	var transform func(dataType *DataSchema)
	transform = func(dataType *DataSchema) {
		if dataType == nil {
			return
		}
		switch dataType.Type {
		case DataTypeString,
			DataTypeNumber,
			DataTypeInteger,
			DataTypeBoolean,
			DataTypeFile:
			// do nothing

		case DataTypeArray:
			transform(dataType.Item)

		case DataTypeObject:
			utils.RangeMapInOrder(
				dataType.Properties,
				func(a, b string) bool { return a < b },
				func(name string, dataSchema *DataSchema) {
					transform(dataSchema)
				},
			)

		default:
			typeArgs = append(typeArgs, dataTypeFallback(dataType))
		}
	}
	transform(d.dataSchema)

	if len(typeArgs) == 0 {
		return genericType
	}
	return spec.NewObjectSchema().
		WithExtendedType(spec.NewSpecificExtendType(genericType, typeArgs...)).
		NewRef()
}

func (d *DataSchemaTransformer) basicSchemaType(s string) *spec.SchemaRef {
	return &spec.Schema{
		Type: s,
	}
}
