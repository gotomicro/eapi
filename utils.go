package eapi

import (
	"go/types"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cast"
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

func ConvertStrToBasicType(str string, t *types.Basic) interface{} {
	switch t.Kind() {
	case types.Bool:
		return cast.ToBool(str)
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
		val, _ := strconv.ParseInt(str, 10, 64)
		return val
	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		val, _ := strconv.ParseUint(str, 10, 64)
		return val
	case types.Float32, types.Float64:
		return cast.ToFloat64(str)
	case types.String:
		return strings.Trim(str, "\"")
	default:
		return str
	}
}
