package spec

import (
	"strings"
)

func RefSchema(id string) *SchemaRef {
	return NewSchemaRef(id, nil)
}

func ArrayProperty(item *SchemaRef) *SchemaRef {
	val := NewArraySchema()
	val.Items = item
	return NewSchemaRef("", val)
}

func MapProperty(value *SchemaRef) *SchemaRef {
	return NewSchemaRef("", NewObjectSchema().WithAdditionalProperties(value.Value))
}

func Unref(t *T, schema *SchemaRef) *SchemaRef {
	if schema.Ref == "" {
		return schema
	}

	slices := strings.Split(schema.Ref, "/")
	if len(slices) != 4 {
		return nil
	}

	switch slices[2] {
	case "schemas":
		return t.Components.Schemas[slices[3]]
	}

	return nil
}
