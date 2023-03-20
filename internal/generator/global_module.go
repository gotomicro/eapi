package generator

var globalModules = map[string]string{} // name -> source code

func LoadGlobalModuleFromSource(name, code string) {
	globalModules[name] = code
}
