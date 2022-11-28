package spec

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func RefSchema(id string) *openapi3.SchemaRef {
	return openapi3.NewSchemaRef(id, nil)
}

func ArrayProperty(item *openapi3.SchemaRef) *openapi3.SchemaRef {
	val := openapi3.NewArraySchema()
	val.Items = item
	return openapi3.NewSchemaRef("", val)
}

func MapProperty(value *openapi3.SchemaRef) *openapi3.SchemaRef {
	return openapi3.NewSchemaRef("", openapi3.NewObjectSchema().WithAdditionalProperties(value.Value))
}

func Unref(t *openapi3.T, schema *openapi3.SchemaRef) *openapi3.SchemaRef {
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
