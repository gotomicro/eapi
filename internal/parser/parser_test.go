package parser

import (
	"testing"
)

func Test_astParser_parserStruct(t *testing.T) {
	AstParserBuild(UserOption{
		RootMainGo: "main.go",
		RootPath:   "testdata/bff",
		GoMod:      "bff",
		RootFile:   "testdata/bff/pkg/router/router_web.go",
	})
}
