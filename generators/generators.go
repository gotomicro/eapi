package generators

import "github.com/getkin/kin-openapi/openapi3"

type Generator struct {
	Type  string
	Items []*Item
}

type Item struct {
	FileName string
	Print    func(schema *openapi3.T) string
}

var Generators = make(map[string]*Generator)

func RegisterGenerator(s *Generator) {
	Generators[s.Type] = s
}
