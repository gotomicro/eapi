package utils

import (
	"go/ast"
	"go/types"
	"os"
	"strings"
)

func Debug() bool {
	return os.Getenv("DEBUG") == "true" || strings.HasSuffix(os.Args[0], ".test")
}

// UnwrapCall unwrap and returns the first nested call
// e.g. unwrapCall(`a(b(c(d)), b1(c1))`) return `c(d)`
func UnwrapCall(callExpr *ast.CallExpr) *ast.CallExpr {
	if len(callExpr.Args) == 0 {
		return callExpr
	}

	arg0 := callExpr.Args[0]
	arg0Call, ok := arg0.(*ast.CallExpr)
	if ok {
		return UnwrapCall(arg0Call)
	}

	return callExpr
}

func GetFuncInfo(fn *types.Func) (typeName, methodName string) {
	sign := fn.Type().(*types.Signature)
	if sign.Recv() != nil {
		typeName = sign.Recv().Type().String()
	} else {
		typeName = fn.Pkg().Path()
	}
	methodName = fn.Name()
	return
}
