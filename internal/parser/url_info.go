package parser

import (
	"github.com/go-openapi/spec"
)

type UrlInfo struct {
	CurrentMod  string
	UniqueKey   string
	FullPath    string
	Method      string
	Prefix      string
	PackagePath string
	ModuleName  string
	FuncName    string
	ReqParam    string
	Swagger     *spec.Schema
}
