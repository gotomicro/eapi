package generators

import "github.com/gotomicro/eapi/spec"

type Generator struct {
	Type  string
	Items []*Item
}

type Item struct {
	FileName string
	Print    func(schema *spec.T) string
}

var Generators = make(map[string]*Generator)

func RegisterGenerator(s *Generator) {
	Generators[s.Type] = s
}
