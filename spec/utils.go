package spec

import (
	"strings"
)

func RefSchema(id string) *SchemaRef {
	return NewSchemaRef(id, nil)
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
		schema = t.Components.Schemas[slices[3]]
		if schema != nil && schema.Ref != "" {
			return Unref(t, schema)
		}
		return schema
	}

	return nil
}

func RefTo(path ...string) *SchemaRef {
	return RefSchema("#/" + strings.Join(path, "/"))
}

func RefComponentSchemas(key string) *SchemaRef {
	return RefTo("components", "schemas", key)
}
