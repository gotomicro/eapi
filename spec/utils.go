package spec

import (
	"strings"
)

func RefSchema(id string) *Schema {
	return &SchemaRef{Ref: id}
}

func Unref(t *T, schema *Schema) *Schema {
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

func RefTo(path ...string) *Schema {
	return RefSchema("#/" + strings.Join(path, "/"))
}

func RefComponentSchemas(key string) *Schema {
	return RefTo("components", "schemas", key)
}
