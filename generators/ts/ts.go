package ts

import (
	_ "embed"
	"fmt"
	"strings"

	f "github.com/gotomicro/eapi/formatter"
	"github.com/gotomicro/eapi/generators"
	"github.com/gotomicro/eapi/spec"
	"github.com/gotomicro/eapi/utils"
	"github.com/samber/lo"
	"github.com/spf13/cast"
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
		Print: func(schema *spec.T, options *generators.PrintOptions) string {
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

func (p *Printer) SetTypeFieldsInline(inline bool) *Printer {
	p.TypeFieldsInLine = true
	return p
}

func (p *Printer) Print() f.Doc {
	var docs []f.Doc
	utils.RangeMapInOrder(
		p.schema.Components.Schemas,
		func(a, b string) bool { return a < b },
		func(key string, schema *spec.SchemaRef) {
			docs = append(docs, p.definition(schema))
		},
	)
	return f.Join(f.Group(f.LineBreak(), f.LineBreak()), docs...)
}

func (p *Printer) definition(definition *spec.SchemaRef) f.Doc {
	if definition.Ref != "" {
		// ignore
		return f.Group()
	}

	ext := definition.Value.ExtendedTypeInfo
	if ext != nil && ext.Type == spec.ExtendedTypeEnum { // enum
		return f.Group(
			f.Content("export enum ", definition.Value.Title, " "),
			p.PrintEnumBody(ext.EnumItems),
		)
	}

	var description string
	if definition.Value != nil {
		description = definition.Value.Description
	}
	if definition.Description != "" {
		description = definition.Description
	}
	description = strings.TrimSpace(description)

	return f.Group(
		f.If(
			description != "",
			p.multilineComment(&multilineCommentOptions{
				tags: []*multilineCommentTag{
					{
						tag:  "@description",
						text: strings.Split(description, "\n\n"),
					},
				},
			}),
		),
		f.Content("export type "+definition.Value.Title+" = "),
		p.PrintType(definition),
	)
}

func (p *Printer) PrintEnumBody(enum []*spec.ExtendedEnumItem) f.Doc {
	return f.Group(
		f.Content("{"), f.LineBreak(),
		f.Indent(f.Group(lo.Map(enum, func(item *spec.ExtendedEnumItem, _ int) f.Doc {
			return f.Group(f.Content(item.Key, " = ", p.printValue(item.Value), ","), f.LineBreak())
		})...)),
		f.Content("}"),
	)
}

func (p *Printer) PrintType(definition *spec.SchemaRef) f.Doc {
	if definition.Ref != "" {
		referencedType := spec.UnrefRecursively(p.schema, definition)
		if referencedType == nil {
			return f.Content("unknown")
		}
		typeName := referencedType.Value.Title
		p.ReferencedTypes = lo.Uniq(append(p.ReferencedTypes, typeName))
		return f.Content(typeName)
	}

	schema := definition.Value
	if schema.ExtendedTypeInfo != nil {
		return p.printExtendedType(schema.ExtendedTypeInfo)
	}

	var t = schema.Type
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
	utils.RangeMapInOrder(
		definition.Value.Properties,
		func(a, b string) bool { return a < b },
		func(name string, property *spec.SchemaRef) {
			required := lo.Contains(definition.Value.Required, name)
			fields = append(fields, p.property(name, property, required))
		},
	)

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
	var description string
	if schema.Value != nil {
		description = schema.Value.Description
	}
	if schema.Description != "" {
		description = schema.Description
	}
	description = strings.TrimSpace(description)

	return f.Group(
		f.If(
			!p.TypeFieldsInLine && description != "",
			p.multilineComment(&multilineCommentOptions{
				tags: []*multilineCommentTag{
					{
						tag:  "@description",
						text: strings.Split(description, "\n"),
					},
				},
			}),
		),
		f.Content(name), f.If(!required, f.Content("?")), f.Content(": "),
		p.PrintType(schema),
		f.Content(";"),
	)
}

type multilineCommentOptions struct {
	// optional
	text string
	// optional
	tags []*multilineCommentTag
}

type multilineCommentTag struct {
	tag  string
	text []string
}

func (p *Printer) multilineComment(options *multilineCommentOptions) f.Doc {
	return f.Group(
		f.Content("/*"), f.LineBreak(),
		f.If(options.text != "", f.Content(" * "+options.text)),
		f.If(len(options.tags) > 0, f.Group(lo.Map(options.tags, func(t *multilineCommentTag, _ int) f.Doc {
			return f.Group(
				f.Content(" * "+t.tag+" "),
				f.Join(f.Group(f.LineBreak(), f.Content(" *\t")), lo.Map(t.text, func(line string, _ int) f.Doc {
					return f.Content(strings.TrimSpace(line))
				})...),
			)
		})...)), f.LineBreak(),
		f.Content(" */"), f.LineBreak(),
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

func (p *Printer) printExtendedType(info *spec.ExtendedTypeInfo) f.Doc {
	switch info.Type {
	case spec.ExtendedTypeAny:
		return f.Content("any")
	case spec.ExtendedTypeMap:
		return f.Content("Record<", p.PrintType(info.Key), ", ", p.PrintType(info.Value), ">")
	}
	return f.Content("unknown")
}

func (p *Printer) printValue(value interface{}) interface{} {
	switch value := value.(type) {
	case string:
		return "\"" + value + "\""
	case int64, uint64:
		return fmt.Sprintf("%d", value)
	case float64:
		return fmt.Sprintf("%f", value)
	case bool:
		if value {
			return "true"
		}
		return "false"
	}
	return cast.ToString(value)
}
