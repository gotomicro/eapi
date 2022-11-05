package analyzer

import (
	"strings"

	"golang.org/x/tools/go/packages"
)

func InspectPackage(pkg *packages.Package, visit func(pkg *packages.Package) bool) {
	if !visit(pkg) {
		return
	}

	for _, p := range pkg.Imports {
		InspectPackage(p, visit)
	}
}

func NormalizeComment(text, trimStart string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, trimStart)
	return text
}
