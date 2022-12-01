package analyzer

import (
	"go/ast"

	"github.com/knadh/koanf"
)

// Plugin 用于对解析逻辑进行扩展以支持不同的框架/模式
type Plugin interface {
	Name() string
	Mount(k *koanf.Koanf) error
	Analyze(ctx *Context, node ast.Node)
}
