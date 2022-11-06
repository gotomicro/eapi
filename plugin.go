package analyzer

import (
	"go/ast"
)

// Plugin 用于对解析逻辑进行扩展以支持不同的框架/模式
type Plugin interface {
	Analyze(ctx *Context, node ast.Node)
}
