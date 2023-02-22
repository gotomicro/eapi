package axios

import (
	"regexp"
	"strings"

	f "github.com/gotomicro/eapi/formatter"
	"github.com/gotomicro/eapi/generators"
	"github.com/gotomicro/eapi/generators/ts"
	"github.com/gotomicro/eapi/spec"
	"github.com/gotomicro/eapi/utils"
	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
)

var (
	Generator = &generators.Generator{
		Type: "axios",
		Items: []*generators.Item{
			ts.TypeGenerator,
			RequestGenerator,
		},
	}

	RequestGenerator = &generators.Item{
		FileName: "api.ts",
		Print: func(schema *spec.T, options *generators.PrintOptions) string {
			return f.Format(NewPrinter(schema, options).Print(), &f.Options{IndentWidth: 2})
		},
	}
)

func init() {
	generators.RegisterGenerator(Generator)
}

type config struct {
	CustomHeader        string
	RemoveFuncPkgPrefix bool
}

type Printer struct {
	schema      *spec.T
	importTypes []string
	options     *generators.PrintOptions
	config
}

func NewPrinter(schema *spec.T, options *generators.PrintOptions) *Printer {
	cfg := config{}
	err := options.ConfigUnmarshaller(&cfg)
	if err != nil {
		panic(err)
	}

	return &Printer{schema: schema, options: options, config: cfg}
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
	if p.CustomHeader != "" {
		return f.Group(f.Content(p.CustomHeader), f.LineBreak())
	}

	return f.Group(
		f.Content(`import axios, { AxiosRequestConfig } from "axios";`),
		f.LineBreak(),
	)
}

type pathItem struct {
	path string
	*spec.PathItem
}

func (p *Printer) requests() f.Doc {
	var docs []f.Doc

	utils.RangeMapInOrder(
		p.schema.Paths,
		func(a, b string) bool { return a < b },
		func(path string, item *spec.PathItem) {
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
		},
	)

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
			if mediaType.Schema.Ref != "" {
				s := spec.Unref(p.schema, mediaType.Schema)
				p.importType(s.Value.Title)
				params = append(params, f.Content("data: "+s.Value.Title))
			} else {
				params = append(params, f.Group(
					f.Content("data: "),
					ts.NewPrinter(p.schema).SetTypeFieldsInline(true).PrintTypeName(mediaType.Schema),
				))
			}
		}
	}
	params = append(params, f.Content("config?: AxiosRequestConfig"))

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
				f.Content("const formData = new FormData();"), f.LineBreak(),
				f.Content("Object.keys(data).forEach((key) => "), f.Content("formData.append(key, data[key])"), f.Content(");"),
				f.LineBreak(),
			))
		}
	}

	var request f.Doc = f.Content("return axios." + method + "(`" + pathName + "`")
	if item.Responses != nil {
		for status := 200; status < 300; status++ {
			response := item.Responses.Get(status)
			if response != nil {
				request = f.Group(f.Content(`return axios.`+method+`<`), p.responseType(response.Value), f.Content(">(`", pathName, "`"))
			}
		}
	}
	res.Docs = append(res.Docs, request)

	options := f.Group()
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
	options.Docs = append(options.Docs, f.LineBreak(), f.Content("...config,"))
	res.Docs = append(res.Docs,
		f.Content(", {"),
		f.Indent(options), f.LineBreak(),
		f.Content("}"),
	)

	res.Docs = append(res.Docs, f.Content(");"))
	return res
}

func (p *Printer) toLowerCamelCase(id string) string {
	return strcase.ToLowerCamel(id)
}

func (p *Printer) paramsType(params []*spec.ParameterRef) f.Doc {
	var fields []f.Doc
	for _, param := range params {
		fields = append(fields, f.Group(
			f.Content(param.Value.Name+"?: "),
			ts.NewPrinter(p.schema).SetTypeFieldsInline(true).PrintTypeName(param.Value.Schema),
		))
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

	lines := strings.Split(desc, "\n\n")
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
		if schema == nil {
			continue
		}
		tsPrinter := ts.NewPrinter(p.schema).SetTypeFieldsInline(true)
		ret := tsPrinter.PrintTypeName(schema)
		p.importType(tsPrinter.ReferencedTypes...)
		return ret
	}

	return f.Content("any")
}

func (p *Printer) requestFnName(item *spec.Operation) string {
	slices := strings.Split(item.OperationID, ".")
	if len(slices) == 1 {
		return p.toLowerCamelCase(item.OperationID)
	}
	if p.RemoveFuncPkgPrefix {
		return p.toLowerCamelCase(slices[len(slices)-1])
	}

	var res = p.toLowerCamelCase(slices[0])
	for _, s := range slices[1:] {
		res += strcase.ToCamel(s)
	}
	return res
}

func (p *Printer) importType(types ...string) {
	p.importTypes = append(p.importTypes, types...)
}