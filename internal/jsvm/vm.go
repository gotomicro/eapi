package jsvm

import (
	_ "embed"
	"fmt"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/gotomicro/eapi/spec"
)

type VM struct {
	vm            *goja.Runtime
	requireModule *require.RequireModule
	registry      *require.Registry
}

func New() *VM {
	vm := &VM{vm: goja.New()}
	vm.init()
	return vm
}

func (v *VM) Run(code string, doc *spec.T) (res string, err error) {
	const mainModule = "__main"
	v.LoadModule(mainModule, code)
	exports, err := v.requireModule.Require(mainModule)
	if err != nil {
		return
	}

	printer, ok := goja.AssertFunction(exports.ToObject(v.vm).Get("print"))
	if !ok {
		err = fmt.Errorf("print is not a function")
		return
	}
	printRes, err := printer(goja.Undefined(), v.vm.ToValue(doc))
	if err != nil {
		return
	}
	res = printRes.String()
	return
}

var (
	//go:embed lib/dist/index.js
	eapiJS string
)

func (v *VM) init() {
	v.vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", false))
	v.registry = &require.Registry{}
	v.requireModule = v.registry.Enable(v.vm)
	console.Enable(v.vm)
	v.LoadModule("eapi", eapiJS)
}

func (v *VM) LoadModule(name, code string) {
	v.registry.RegisterNativeModule(name, func(runtime *goja.Runtime, module *goja.Object) {
		f, err := runtime.RunString("(function (module, exports, require) {" + code + "})")
		if err != nil {
			panic(err)
		}
		fn, ok := goja.AssertFunction(f)
		if !ok {
			panic("invalid module " + name)
		}
		exports := module.Get("exports")
		_, err = fn(exports, module, exports, runtime.Get("require"))
		if err != nil {
			panic(err)
		}
	})
}

func (v *VM) Require(module string) (goja.Value, error) {
	return v.requireModule.Require(module)
}

func (v *VM) VM() *goja.Runtime {
	return v.vm
}
