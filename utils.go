package eapi

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
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

func ReadGoMod(pkgPath string) (mod *modfile.File, err error) {
	fileName := filepath.Join(pkgPath, "go.mod")
	content, err := os.ReadFile(fileName)
	if err != nil {
		return
	}

	mod, err = modfile.Parse("go.mod", content, nil)
	if err != nil {
		return
	}

	return
}
