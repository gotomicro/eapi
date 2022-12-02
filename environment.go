package eapi

type Environment struct {
	parent  *Environment
	records map[interface{}]interface{}
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{parent: parent, records: make(map[interface{}]interface{})}
}

func (e *Environment) Define(k, v interface{}) *Environment {
	e.records[k] = v
	return e
}

func (e *Environment) Lookup(k interface{}) interface{} {
	env := e.Resolve(k)
	if env == nil {
		return nil
	}
	return env.records[k]
}

func (e *Environment) Resolve(k interface{}) *Environment {
	_, ok := e.records[k]
	if ok {
		return e
	}
	if e.parent != nil {
		return e.parent.Resolve(k)
	}

	return nil
}

func (e *Environment) Assign(k, v interface{}) *Environment {
	scope := e.Resolve(k)
	if scope != nil {
		scope.Define(k, v)
	}
	return e
}
