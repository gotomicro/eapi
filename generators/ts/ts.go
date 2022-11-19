package ts

import (
	_ "embed"

	"github.com/go-openapi/spec"
	f "github.com/gotomicro/ego-gen-api/formatter"
	"github.com/gotomicro/ego-gen-api/generators"
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
		Print: func(schema *spec.Swagger) string {
			return f.Format(NewPrinter(schema).Print(), &f.Options{IndentWidth: 2})
		},
	}
)

func init() {
	generators.RegisterGenerator(Generator)
}

type Printer struct {
	schema *spec.Swagger
}

func NewPrinter(schema *spec.Swagger) *Printer {
	return &Printer{schema: schema}
}

func (p *Printer) Print() f.Doc {
	var docs []f.Doc
	for _, definition := range p.schema.Definitions {
		docs = append(docs, p.definition(definition))
	}
	return f.Join(f.Group(f.LineBreak(), f.LineBreak()), docs...)
}

func (p *Printer) definition(definition spec.Schema) f.Doc {
	return f.Group(
		f.Content("export type "+definition.Title+" = "),
		p.printType(definition),
	)
}

func (p *Printer) printType(definition spec.Schema) f.Doc {
	var t = "object"
	if len(definition.Type) > 0 {
		t = definition.Type[0]
	}

	if !definition.Ref.GetPointer().IsEmpty() {
		tokens := definition.Ref.GetPointer().DecodedTokens()
		if len(tokens) != 2 || tokens[0] != "definitions" {
			return f.Content("unknown")
		}
		return f.Content(p.schema.Definitions[tokens[1]].Title)
	}

	switch t {
	case "object":
		return p.printInterface(definition)
	case "array":
		schema := definition.Items.Schema
		if schema == nil {
			return f.Content("any[]")
		}
		return f.Group(
			p.printType(*schema),
			f.Content("[]"),
		)
	default:
		return p.printBasicType(t)
	}
}

func (p *Printer) printInterface(definition spec.Schema) f.Doc {
	var fields []f.Doc
	for name, schema := range definition.Properties {
		required := lo.Contains(definition.Required, name)
		fields = append(fields, p.property(name, schema, required))
	}

	return f.Group(
		f.Content("{"),
		f.LineBreak(),
		f.Indent(f.Join(f.LineBreak(), fields...)),
		f.LineBreak(),
		f.Content("}"),
	)
}

func (p *Printer) property(name string, schema spec.Schema, required bool) f.Doc {
	var content = name
	if !required {
		content += "?"
	}
	content += ": "

	return f.Group(
		f.Content(content),
		p.printType(schema),
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
	return f.Content("unknown")
}
