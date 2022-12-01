package ts

import (
	_ "embed"

	f "github.com/gotomicro/ego-gen-api/formatter"
	"github.com/gotomicro/ego-gen-api/generators"
	"github.com/gotomicro/ego-gen-api/spec"
	"github.com/samber/lo"
)

var (
	// Generator for typescript types
	Generator = &generators.Generator{
		Type: "ts",
		Items: []*generators.Item{
			TypeGenerator,
		},
	}

	TypeGenerator = &generators.Item{
		FileName: "types.ts",
		Print: func(schema *spec.T) string {
			return f.Format(NewPrinter(schema).Print(), &f.Options{IndentWidth: 2})
		},
	}
)

func init() {
	generators.RegisterGenerator(Generator)
}

type Printer struct {
	schema *spec.T

	ReferencedTypes []string
	// 类型的字段是否在一行
	TypeFieldsInLine bool
}

func NewPrinter(schema *spec.T) *Printer {
	return &Printer{schema: schema}
}

func (p *Printer) Print() f.Doc {
	var docs []f.Doc
	for _, schema := range p.schema.Components.Schemas {
		docs = append(docs, p.definition(schema))
	}
	return f.Join(f.Group(f.LineBreak(), f.LineBreak()), docs...)
}

func (p *Printer) definition(definition *spec.SchemaRef) f.Doc {
	return f.Group(
		f.Content("export type "+definition.Value.Title+" = "),
		p.PrintType(definition),
	)
}

func (p *Printer) PrintType(definition *spec.SchemaRef) f.Doc {
	if definition.Ref != "" {
		referencedType := spec.Unref(p.schema, definition)
		if referencedType == nil {
			return f.Content("unknown")
		}
		typeName := referencedType.Value.Title
		p.ReferencedTypes = append(p.ReferencedTypes, typeName)
		return f.Content(typeName)
	}

	var t = definition.Value.Type
	switch t {
	case "object":
		return p.printInterface(definition)
	case "array":
		if definition.Value.Items == nil {
			return f.Content("any[]")
		}
		schema := definition.Value.Items
		return f.Group(
			p.PrintType(schema),
			f.Content("[]"),
		)
	default:
		return p.printBasicType(t)
	}
}

func (p *Printer) printInterface(definition *spec.SchemaRef) f.Doc {
	var fields []f.Doc
	for name, schema := range definition.Value.Properties {
		required := lo.Contains(definition.Value.Required, name)
		fields = append(fields, p.property(name, schema, required))
	}

	if p.TypeFieldsInLine {
		return f.Group(
			f.Content("{ "),
			f.Indent(f.Join(f.Content(" "), fields...)),
			f.Content(" }"),
		)
	}

	return f.Group(
		f.Content("{"),
		f.LineBreak(),
		f.Indent(f.Join(f.LineBreak(), fields...)),
		f.LineBreak(),
		f.Content("}"),
	)
}

func (p *Printer) property(name string, schema *spec.SchemaRef, required bool) f.Doc {
	var content = name
	if !required {
		content += "?"
	}
	content += ": "

	return f.Group(
		f.Content(content),
		p.PrintType(schema),
		f.Content(";"),
	)
}

func (p *Printer) printBasicType(t string) f.Doc {
	switch t {
	case "string":
		return f.Content("string")
	case "number", "integer":
		return f.Content("number")
	case "boolean":
		return f.Content("boolean")
	case "file":
		return f.Content("File")
	}
	return f.Content("any")
}
