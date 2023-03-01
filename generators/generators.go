package generators

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/gotomicro/eapi/internal/generator"
	"github.com/gotomicro/eapi/spec"
)

type Generator struct {
	Name  string
	Print func(schema *spec.T, options *PrintOptions) []*generator.GenerateResultItem
}

type PrintOptions struct {
	GetConfig func(key string) interface{}
}

var Generators = make(map[string]*Generator)

func RegisterGenerator(s *Generator) {
	Generators[s.Name] = s
}

func NewGeneratorFromSourceCode(name, code string) *Generator {
	return &Generator{
		Name: name,
		Print: func(schema *spec.T, options *PrintOptions) []*generator.GenerateResultItem {
			res, err := generator.New(options.GetConfig).Run(code, schema)
			if err != nil {
				fmt.Fprintf(os.Stderr, "generator '%s' occurs error. error: %s", name, err.Error())
				return nil
			}
			return res
		},
	}
}

func NewGeneratorFromFile(file string) *Generator {
	return &Generator{
		Name: file,
		Print: func(schema *spec.T, options *PrintOptions) []*generator.GenerateResultItem {
			res, err := generator.New(options.GetConfig).RunFromModule(file, schema)
			if err != nil {
				fmt.Fprintf(os.Stderr, "generator '%s' occurs error. error: %s", file, err.Error())
				return nil
			}
			return res
		},
	}
}

var (
	//go:embed lib/axios.js
	axios string
	//go:embed lib/ts.js
	ts string
	//go:embed lib/umi.js
	umi string
)

func init() {
	generator.VM.LoadModule("eapi/generators/ts", ts)
	generator.VM.LoadModule("eapi/generators/axios", axios)
	generator.VM.LoadModule("eapi/generators/umi", umi)

	RegisterGenerator(NewGeneratorFromSourceCode("axios", axios))
	RegisterGenerator(NewGeneratorFromSourceCode("ts", ts))
	RegisterGenerator(NewGeneratorFromSourceCode("umi", umi))
}
