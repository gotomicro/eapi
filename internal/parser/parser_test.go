package parser

import (
	"testing"
)

func Test_astParser_parserStruct(t *testing.T) {
	AstParserBuild(UserOption{
		GoMod:     "bff",
		RouterMod: "bff/pkg/router",
		RootFile:  "testdata/bff/pkg/router/router_web.go",
	})
}
