package parser

import (
	"testing"
)

func Test_astParser_parserStruct(t *testing.T) {
	AstParserBuild(UserOption{
		GoMod:     "bff",
		RouterMod: "bff/testdata",
		RootFile:  "testdata/router_web.go",
	})
}
