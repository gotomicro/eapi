// Package umi
// 生成使用 umi-request 进行接口请求的代码及请求、相应参数的 TS 类型生命
package umi

import (
	"regexp"
	"sort"
	"strings"

	f "github.com/gotomicro/ego-gen-api/formatter"
	"github.com/gotomicro/ego-gen-api/generators"
	"github.com/gotomicro/ego-gen-api/generators/ts"
	"github.com/gotomicro/ego-gen-api/spec"
	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
)

var (
	Generator = &generators.Generator{
		Type: "umi",
		Items: []*generators.Item{
			ts.TypeGenerator,
			RequestGenerator,
		},
	}

	RequestGenerator = &generators.Item{
		FileName: "request.ts",
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

	importTypes []string
}

func NewPrinter(schema *spec.T) *Printer {
	return &Printer{schema: schema}
}

func (p *Printer) Print() f.Doc {
	var request = p.requests()
	return f.Group(
		p.header(),
		p.imports(),
		f.LineBreak(),
		request,
	)
}

func (p *Printer) header() f.Doc {
	return f.Group(
		f.Content(`import { request } from "umi";`),
		f.LineBreak(),
	)
}

type pathItem struct {
	path string
	*spec.PathItem
}

func (p *Printer) requests() f.Doc {
	var docs []f.Doc
	var paths []pathItem
	for path, item := range p.schema.Paths {
		paths = append(paths, pathItem{path: path, PathItem: item})
	}
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].path < paths[j].path
	})

	for _, item := range paths {
		path := item.path
		if item.Get != nil {
			docs = append(docs, p.request(path, "get", item.Get))
		}
		if item.Put != nil {
			docs = append(docs, p.request(path, "put", item.Put))
		}
		if item.Post != nil {
			docs = append(docs, p.request(path, "post", item.Post))
		}
		if item.Delete != nil {
			docs = append(docs, p.request(path, "delete", item.Delete))
		}
		if item.Options != nil {
			docs = append(docs, p.request(path, "options", item.Options))
		}
		if item.Head != nil {
			docs = append(docs, p.request(path, "head", item.Head))
		}
		if item.Patch != nil {
			docs = append(docs, p.request(path, "patch", item.Patch))
		}
	}

	return f.Join(f.Group(f.LineBreak(), f.LineBreak()), docs...)
}

var pathParamPattern = regexp.MustCompile("\\{([\\w-]+)\\}")

func (p *Printer) request(path string, method string, item *spec.Operation) f.Doc {
	var params []f.Doc
	var queryParams []*spec.ParameterRef
	var pathParams []*spec.ParameterRef
	for _, parameter := range item.Parameters {
		p := parameter
		switch parameter.Value.In {
		case "path":
			pathParams = append(pathParams, p)
		case "query":
			queryParams = append(queryParams, p)
		}
	}

	pathName := strings.ReplaceAll(path, "{", "${")
	matches := pathParamPattern.FindAllStringSubmatch(path, -1)
	for _, match := range matches {
		originalName := match[1]
		name := p.toLowerCamelCase(originalName)
		if name != originalName {
			pathName = strings.ReplaceAll(pathName, "{"+originalName+"}", "{"+name+"}")
			pathParams = lo.Filter(pathParams, func(p *spec.ParameterRef, _ int) bool { return p.Value.Name != originalName })
		}
		exists := lo.ContainsBy(pathParams, func(p *spec.ParameterRef) bool { return p.Value.Name == name })
		if !exists {
			pathParams = append(pathParams, &spec.ParameterRef{Value: spec.NewPathParameter(name)})
		}
	}

	if len(pathParams) > 0 {
		for _, param := range pathParams {
			params = append(params, f.Group(f.Content(param.Value.Name+": string")))
		}
	}
	if len(queryParams) > 0 {
		params = append(params, f.Group(f.Content("query: "), p.paramsType(queryParams)))
	}

	if item.RequestBody != nil {
		_, mediaType := p.getRequestMediaType(item)
		if mediaType != nil {
			s := spec.Unref(p.schema, mediaType.Schema)
			p.importTypes = append(p.importTypes, s.Value.Title)
			params = append(params, f.Content("data: "+s.Value.Title))
		}
	}

	functionBody := p.requestFunctionBody(pathName, method, queryParams, item)
	return f.Group(
		p.jsDoc(item),
		f.Content(`export function `+p.requestFnName(item)+"("),
		f.Join(f.Content(", "), params...),
		f.Content(") {"),
		f.LineBreak(),
		f.Indent(functionBody),
		f.LineBreak(),
		f.Content("}"),
	)
}

const contentTypeJson = "application/json"
const contentTypeFormData = "multipart/form-data"

var (
	contentTypeInPriority = []string{
		"application/json",
		"multipart/form-data",
	}
)

func (p *Printer) getRequestMediaType(item *spec.Operation) (string, *spec.MediaType) {
	if item.RequestBody == nil || item.RequestBody.Value == nil || item.RequestBody.Value.Content == nil {
		return "", nil
	}

	for _, contentType := range contentTypeInPriority {
		res := item.RequestBody.Value.GetMediaType(contentType)
		if res != nil {
			return contentType, res
		}
	}

	for contentType, mediaType := range item.RequestBody.Value.Content {
		return contentType, mediaType
	}

	return "", nil
}

func (p *Printer) requestFunctionBody(pathName string, method string, queryParams []*spec.ParameterRef, item *spec.Operation) *f.DocGroup {
	res := f.Group()

	reqContentType, mediaType := p.getRequestMediaType(item)
	if mediaType != nil {
		if reqContentType == contentTypeFormData {
			res.Docs = append(res.Docs, f.Group(
				f.Content("let formData = new FormData();"), f.LineBreak(),
				f.Content("for (const key in data) {"), f.LineBreak(),
				f.Indent(f.Group(
					f.Content("formData.append(key, data[key as keyof typeof data] as string);"), f.LineBreak(),
				)),
				f.Content("}"),
				f.LineBreak(),
			))
		}
	}

	var request f.Doc = f.Content("return request(`" + pathName + "`, {")
	if item.Responses != nil {
		for status := 200; status < 300; status++ {
			response := item.Responses.Get(status)
			if response != nil {
				request = f.Group(f.Content("return request<"), p.responseType(response.Value), f.Content(">(`"+pathName+"`, {"))
			}
		}
	}
	res.Docs = append(res.Docs, request)

	options := f.Group(f.LineBreak(), f.Content(`method: "`+method+`",`))
	if len(queryParams) > 0 {
		options.Docs = append(options.Docs, f.LineBreak(), f.Content("params: query,"))
	}
	if mediaType != nil {
		if reqContentType == contentTypeFormData {
			options.Docs = append(options.Docs, f.LineBreak(), f.Content("data: formData,"))
		} else {
			options.Docs = append(options.Docs, f.LineBreak(), f.Content("data,"))
		}
	}
	res.Docs = append(res.Docs, f.Indent(options))
	res.Docs = append(res.Docs, f.LineBreak(), f.Content("}"), f.Content(");"))
	return res
}

func (p *Printer) toLowerCamelCase(id string) string {
	return strcase.ToLowerCamel(id)
}

func (p *Printer) paramsType(params []*spec.ParameterRef) f.Doc {
	var fields []f.Doc
	for _, param := range params {
		typeName := param.Value.Schema.Value.Type
		switch typeName {
		case "integer":
			typeName = "number"
		case "array":
			var itemType string
			items := param.Value.Schema.Value.Items
			if items != nil && items.Value != nil {
				itemType = items.Value.Type
			}
			if itemType == "" {
				itemType = "string"
			}
			typeName = itemType + "[]"

		case "":
			typeName = "string"
		}
		fields = append(fields, f.Content(param.Value.Name+": "+typeName))
	}

	return f.Group(
		f.Content("{ "),
		f.Join(f.Content("; "), fields...),
		f.Content(" }"),
	)
}

func (p *Printer) imports() f.Doc {
	if len(p.importTypes) == 0 {
		return f.Group()
	}
	p.importTypes = lo.Uniq(p.importTypes)

	var types []f.Doc
	var length = len("import { ")
	for i, importType := range p.importTypes {
		if i > 0 {
			length += 2
		}
		length += len(importType)
		types = append(types, f.Content(importType))
	}
	length += len(" } from \"./types\";")
	if length > 80 {
		return f.Group(
			f.Content("import { "),
			f.LineBreak(),
			f.Indent(f.Join(f.Group(f.Content(","), f.LineBreak()), types...)),
			f.LineBreak(),
			f.Content(" } from \"./types\";"),
			f.LineBreak(),
		)
	}

	return f.Group(
		f.Content("import { "),
		f.Join(f.Content(", "), types...),
		f.Content(" } from \"./types\";"),
		f.LineBreak(),
	)
}

func (p *Printer) jsDoc(item *spec.Operation) f.Doc {
	desc := strings.TrimSpace(item.Description)
	if desc == "" {
		return f.Group()
	}

	lines := strings.Split(desc, "\n")
	var res = f.Group()
	res.Docs = append(res.Docs,
		f.Content("/*"), f.LineBreak(),
		f.Content(" * @description "+strings.TrimSpace(lines[0])), f.LineBreak(),
	)
	for _, line := range lines[1:] {
		res.Docs = append(res.Docs, f.Content(" * "+strings.TrimSpace(line)), f.LineBreak())
	}
	res.Docs = append(res.Docs, f.Content(" */"), f.LineBreak())
	return res
}

func (p *Printer) responseType(res *spec.Response) f.Doc {
	for _, mediaType := range res.Content {
		schema := mediaType.Schema
		schema = spec.Unref(p.schema, schema)
		tsPrinter := ts.NewPrinter(p.schema)
		tsPrinter.TypeFieldsInLine = true
		ret := tsPrinter.PrintType(schema)
		p.importTypes = append(p.importTypes, tsPrinter.ReferencedTypes...)
		return ret
	}

	return f.Content("any")
}

func (p *Printer) requestFnName(item *spec.Operation) string {
	slices := strings.Split(item.OperationID, ".")
	if len(slices) == 1 {
		return p.toLowerCamelCase(item.OperationID)
	}

	var res = p.toLowerCamelCase(slices[0])
	for _, s := range slices[1:] {
		res += strcase.ToCamel(s)
	}
	return res
}
