package parser

import (
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/loader"
)

// PackagesDefinitions map[package import path]*PackageDefinitions.
type PackagesDefinitions struct {
	files             map[*ast.File]*AstFileInfo
	packages          map[string]*PackageDefinitions
	uniqueDefinitions map[string]*TypeSpecDef
}

// NewPackagesDefinitions create object PackagesDefinitions.
func NewPackagesDefinitions() *PackagesDefinitions {
	return &PackagesDefinitions{
		files:             make(map[*ast.File]*AstFileInfo),
		packages:          make(map[string]*PackageDefinitions),
		uniqueDefinitions: make(map[string]*TypeSpecDef),
	}
}

// CollectAstFile collect ast.file.
func (pkgDefs *PackagesDefinitions) CollectAstFile(packageDir, path string, astFile *ast.File) error {
	if pkgDefs.files == nil {
		pkgDefs.files = make(map[*ast.File]*AstFileInfo)
	}

	if pkgDefs.packages == nil {
		pkgDefs.packages = make(map[string]*PackageDefinitions)
	}
	// bff/pkg/shop
	fmt.Printf("packageDir--------------->"+"%+v\n", packageDir)
	// testdata/bff/pkg/shop/shop.go
	fmt.Printf("path--------------->"+"%+v\n", path)
	// return without storing the file if we lack a packageDir
	if packageDir == "" {
		return nil
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	dependency, ok := pkgDefs.packages[packageDir]
	if ok {
		// return without storing the file if it already exists
		_, exists := dependency.Files[path]
		if exists {
			return nil
		}

		dependency.Files[path] = astFile
	} else {
		pkgDefs.packages[packageDir] = &PackageDefinitions{
			Name:            astFile.Name.Name,
			Files:           map[string]*ast.File{path: astFile},
			TypeDefinitions: make(map[string]*TypeSpecDef),
		}
	}

	pkgDefs.files[astFile] = &AstFileInfo{
		File:        astFile,
		Path:        path,
		PackagePath: packageDir,
	}

	return nil
}

// ParseTypes 解析定义的结构体
// @Return parsed definitions.
func (pkgDefs *PackagesDefinitions) ParseTypes() (map[*TypeSpecDef]*Schema, error) {
	parsedSchemas := make(map[*TypeSpecDef]*Schema)
	for astFile, info := range pkgDefs.files {
		pkgDefs.parseTypesFromFile(astFile, info.PackagePath, parsedSchemas)
	}
	return parsedSchemas, nil
}

func (pkgDefs *PackagesDefinitions) parseTypesFromFile(astFile *ast.File, packagePath string, parsedSchemas map[*TypeSpecDef]*Schema) {
	for _, astDeclaration := range astFile.Decls {
		if generalDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && generalDeclaration.Tok == token.TYPE {
			for _, astSpec := range generalDeclaration.Specs {
				if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
					typeSpecDef := &TypeSpecDef{
						PkgPath:  packagePath,
						File:     astFile,
						TypeSpec: typeSpec,
					}

					if idt, ok := typeSpec.Type.(*ast.Ident); ok && IsGolangPrimitiveType(idt.Name) && parsedSchemas != nil {
						parsedSchemas[typeSpecDef] = &Schema{
							PkgPath: typeSpecDef.PkgPath,
							Name:    astFile.Name.Name,
							Schema:  PrimitiveSchema(TransToValidSchemeType(idt.Name)),
						}
					}

					if pkgDefs.uniqueDefinitions == nil {
						pkgDefs.uniqueDefinitions = make(map[string]*TypeSpecDef)
					}

					fullName := typeSpecDef.FullName()
					anotherTypeDef, ok := pkgDefs.uniqueDefinitions[fullName]
					if ok {
						if typeSpecDef.PkgPath == anotherTypeDef.PkgPath {
							continue
						} else {
							delete(pkgDefs.uniqueDefinitions, fullName)
						}
					} else {
						pkgDefs.uniqueDefinitions[fullName] = typeSpecDef
					}

					if pkgDefs.packages[typeSpecDef.PkgPath] == nil {
						pkgDefs.packages[typeSpecDef.PkgPath] = &PackageDefinitions{
							Name:            astFile.Name.Name,
							TypeDefinitions: map[string]*TypeSpecDef{typeSpecDef.Name(): typeSpecDef},
						}
					} else if _, ok = pkgDefs.packages[typeSpecDef.PkgPath].TypeDefinitions[typeSpecDef.Name()]; !ok {
						pkgDefs.packages[typeSpecDef.PkgPath].TypeDefinitions[typeSpecDef.Name()] = typeSpecDef
					}
				}
			}
		}
	}
}

// 根据 pkg path
// 路径，找到对应的文件
func (pkgDefs *PackagesDefinitions) findTypeSpec(pkgPath string, typeName string) *TypeSpecDef {
	if pkgDefs.packages == nil {
		return nil
	}

	pd, found := pkgDefs.packages[pkgPath]
	if found {
		typeSpec, ok := pd.TypeDefinitions[typeName]
		if ok {
			return typeSpec
		}
	}

	return nil
}

func (pkgDefs *PackagesDefinitions) findFileByRangeFiles(handle func(filename string, file *ast.File) (bool, error)) (file *ast.File, err error) {
	if pkgDefs.files == nil {
		return nil, fmt.Errorf("files is empty")
	}

	for file, info := range pkgDefs.files {
		flag, err := handle(info.Path, file)
		if err != nil {
			return nil, fmt.Errorf("handle file %s error: %w", info.Path, err)
		}
		// 说明找到了
		if !flag {
			return file, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

func (pkgDefs *PackagesDefinitions) findFileByRangePackages(pkgPath string, handle func(filename string, file *ast.File) (bool, error)) (file *ast.File, err error) {
	if pkgDefs.packages == nil {
		return nil, fmt.Errorf("files is empty")
	}
	pd, found := pkgDefs.packages[pkgPath]
	if !found {
		return nil, fmt.Errorf("not packages found")
	}

	for key, file := range pd.Files {
		flag, err := handle(key, file)
		if err != nil {
			return nil, fmt.Errorf("handle file %s error: %w", key, err)
		}
		// 说明找到了
		if !flag {
			return file, nil
		}
	}

	return nil, fmt.Errorf("not file found")
}

// rangeByPkgPath 根据路径遍历，使用函数处理
func (pkgDefs *PackagesDefinitions) rangeByPkgPath(pkgPath string, handle func(filename string, file *ast.File) error) error {
	if pkgDefs.packages == nil {
		return nil
	}
	pd, found := pkgDefs.packages[pkgPath]
	if !found {
		return nil
	}

	for key, info := range pd.Files {
		err := handle(key, info)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pkgDefs *PackagesDefinitions) loadExternalPackage(importPath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	conf := loader.Config{
		ParserMode: goparser.ParseComments,
		Cwd:        cwd,
	}

	conf.Import(importPath)

	//loaderProgram, err := conf.Load()
	//if err != nil {
	//	return err
	//}

	//for _, info := range loaderProgram.AllPackages {
	//	pkgPath := strings.TrimPrefix(info.Pkg.Path(), "vendor/")
	//	for _, astFile := range info.Files {
	//		pkgDefs.parseTypesFromFile(astFile, pkgPath, nil)
	//	}
	//}

	return nil
}

// findPackagePathFromImports finds out the package path of a package via ranging imports of an ast.File
// @pkg the name of the target package
// @file current ast.File in which to search imports
// @fuzzy search for the package path that the last part matches the @pkg if true
// @return the package path of a package of @pkg.
func (pkgDefs *PackagesDefinitions) findPackagePathFromImports(pkg string, file *ast.File, fuzzy bool) string {
	if file == nil {
		return ""
	}

	if strings.ContainsRune(pkg, '.') {
		pkg = strings.Split(pkg, ".")[0]
	}

	hasAnonymousPkg := false

	matchLastPathPart := func(pkgPath string) bool {
		paths := strings.Split(pkgPath, "/")

		return paths[len(paths)-1] == pkg
	}

	// prior to match named package
	for _, imp := range file.Imports {
		if imp.Name != nil {
			if imp.Name.Name == pkg {
				return strings.Trim(imp.Path.Value, `"`)
			}

			if imp.Name.Name == "_" {
				hasAnonymousPkg = true
			}

			continue
		}

		if pkgDefs.packages != nil {
			path := strings.Trim(imp.Path.Value, `"`)
			if fuzzy {
				if matchLastPathPart(path) {
					return path
				}

				continue
			}

			pd, ok := pkgDefs.packages[path]
			if ok && pd.Name == pkg {
				return path
			}
		}
	}

	// match unnamed package
	if hasAnonymousPkg && pkgDefs.packages != nil {
		for _, imp := range file.Imports {
			if imp.Name == nil {
				continue
			}
			if imp.Name.Name == "_" {
				path := strings.Trim(imp.Path.Value, `"`)
				if fuzzy {
					if matchLastPathPart(path) {
						return path
					}
				} else if pd, ok := pkgDefs.packages[path]; ok && pd.Name == pkg {
					return path
				}
			}
		}
	}

	return ""
}

// FindTypeSpec finds out TypeSpecDef of a type by typeName
// @typeName the name of the target type, if it starts with a package name, find its own package path from imports on top of @file
// @file the ast.file in which @typeName is used
// @pkgPath the package path of @file.
func (pkgDefs *PackagesDefinitions) FindTypeSpec(typeName string, file *ast.File, parseDependency bool) *TypeSpecDef {
	if IsGolangPrimitiveType(typeName) {
		return nil
	}

	if file == nil { // for test
		return pkgDefs.uniqueDefinitions[typeName]
	}

	parts := strings.Split(typeName, ".")
	if len(parts) > 1 {
		isAliasPkgName := func(file *ast.File, pkgName string) bool {
			if file != nil && file.Imports != nil {
				for _, pkg := range file.Imports {
					if pkg.Name != nil && pkg.Name.Name == pkgName {
						return true
					}
				}
			}

			return false
		}

		if !isAliasPkgName(file, parts[0]) {
			typeDef, ok := pkgDefs.uniqueDefinitions[typeName]
			if ok {
				return typeDef
			}
		}

		pkgPath := pkgDefs.findPackagePathFromImports(parts[0], file, false)
		if len(pkgPath) == 0 {
			// check if the current package
			if parts[0] == file.Name.Name {
				pkgPath = pkgDefs.files[file].PackagePath
			} else if parseDependency {
				// take it as an external package, needs to be loaded
				if pkgPath = pkgDefs.findPackagePathFromImports(parts[0], file, true); len(pkgPath) > 0 {
					if err := pkgDefs.loadExternalPackage(pkgPath); err != nil {
						return nil
					}
				}
			}
		}

		return pkgDefs.findTypeSpec(pkgPath, parts[1])
	}

	typeDef, ok := pkgDefs.uniqueDefinitions[fullTypeName(file.Name.Name, typeName)]
	if ok {
		return typeDef
	}

	typeDef = pkgDefs.findTypeSpec(pkgDefs.files[file].PackagePath, typeName)
	if typeDef != nil {
		return typeDef
	}

	for _, imp := range file.Imports {
		if imp.Name != nil && imp.Name.Name == "." {
			typeDef := pkgDefs.findTypeSpec(strings.Trim(imp.Path.Value, `"`), typeName)
			if typeDef != nil {
				return typeDef
			}
		}
	}

	return nil
}

/*
var req GoodCreateReq
req2 := GoodCreateReq{}
var req3 dto.Good3Req
req4 := dto.Good3Req{}

Body: *ast.BlockStmt {
  1056  .  .  .  .  Lbrace: foo:47:33
  1057  .  .  .  .  List: []ast.Stmt (len = 6) {
  1058  .  .  .  .  .  0: *ast.DeclStmt {
  1059  .  .  .  .  .  .  Decl: *ast.GenDecl {
  1060  .  .  .  .  .  .  .  Doc: nil
  1061  .  .  .  .  .  .  .  TokPos: foo:48:2
  1062  .  .  .  .  .  .  .  Tok: var
  1063  .  .  .  .  .  .  .  Lparen: -
  1064  .  .  .  .  .  .  .  Specs: []ast.Spec (len = 1) {
  1065  .  .  .  .  .  .  .  .  0: *ast.ValueSpec {
  1066  .  .  .  .  .  .  .  .  .  Doc: nil
  1067  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
  1068  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
  1069  .  .  .  .  .  .  .  .  .  .  .  NamePos: foo:48:6
  1070  .  .  .  .  .  .  .  .  .  .  .  Name: "req"
  1071  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
  1072  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
  1073  .  .  .  .  .  .  .  .  .  .  .  .  Name: "req"
  1074  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 1065)
  1075  .  .  .  .  .  .  .  .  .  .  .  .  Data: 0
  1076  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
  1077  .  .  .  .  .  .  .  .  .  .  .  }
  1078  .  .  .  .  .  .  .  .  .  .  }
  1079  .  .  .  .  .  .  .  .  .  }
  1080  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
  1081  .  .  .  .  .  .  .  .  .  .  NamePos: foo:48:10
  1082  .  .  .  .  .  .  .  .  .  .  Name: "GoodCreateReq"
  1083  .  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 55)
  1084  .  .  .  .  .  .  .  .  .  }
  1085  .  .  .  .  .  .  .  .  .  Values: nil
  1086  .  .  .  .  .  .  .  .  .  Comment: nil
  1087  .  .  .  .  .  .  .  .  }
  1088  .  .  .  .  .  .  .  }
  1089  .  .  .  .  .  .  .  Rparen: -
  1090  .  .  .  .  .  .  }
  1091  .  .  .  .  .  }
  1092  .  .  .  .  .  1: *ast.AssignStmt {
  1093  .  .  .  .  .  .  Lhs: []ast.Expr (len = 1) {
  1094  .  .  .  .  .  .  .  0: *ast.Ident {
  1095  .  .  .  .  .  .  .  .  NamePos: foo:49:9
  1096  .  .  .  .  .  .  .  .  Name: "req2"
  1097  .  .  .  .  .  .  .  .  Obj: *ast.Object {
  1098  .  .  .  .  .  .  .  .  .  Kind: var
  1099  .  .  .  .  .  .  .  .  .  Name: "req2"
  1100  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 1092)
  1101  .  .  .  .  .  .  .  .  .  Data: nil
  1102  .  .  .  .  .  .  .  .  .  Type: nil
  1103  .  .  .  .  .  .  .  .  }
  1104  .  .  .  .  .  .  .  }
  1105  .  .  .  .  .  .  }
  1106  .  .  .  .  .  .  TokPos: foo:49:14
  1107  .  .  .  .  .  .  Tok: :=
  1108  .  .  .  .  .  .  Rhs: []ast.Expr (len = 1) {
  1109  .  .  .  .  .  .  .  0: *ast.CompositeLit {
  1110  .  .  .  .  .  .  .  .  Type: *ast.Ident {
  1111  .  .  .  .  .  .  .  .  .  NamePos: foo:49:17
  1112  .  .  .  .  .  .  .  .  .  Name: "GoodCreateReq"
  1113  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 55)
  1114  .  .  .  .  .  .  .  .  }
  1115  .  .  .  .  .  .  .  .  Lbrace: foo:49:30
  1116  .  .  .  .  .  .  .  .  Elts: nil
  1117  .  .  .  .  .  .  .  .  Rbrace: foo:49:31
  1118  .  .  .  .  .  .  .  .  Incomplete: false
  1119  .  .  .  .  .  .  .  }
  1120  .  .  .  .  .  .  }
  1121  .  .  .  .  .  }
  1122  .  .  .  .  .  2: *ast.DeclStmt {
  1123  .  .  .  .  .  .  Decl: *ast.GenDecl {
  1124  .  .  .  .  .  .  .  Doc: nil
  1125  .  .  .  .  .  .  .  TokPos: foo:50:9
  1126  .  .  .  .  .  .  .  Tok: var
  1127  .  .  .  .  .  .  .  Lparen: -
  1128  .  .  .  .  .  .  .  Specs: []ast.Spec (len = 1) {
  1129  .  .  .  .  .  .  .  .  0: *ast.ValueSpec {
  1130  .  .  .  .  .  .  .  .  .  Doc: nil
  1131  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
  1132  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
  1133  .  .  .  .  .  .  .  .  .  .  .  NamePos: foo:50:13
  1134  .  .  .  .  .  .  .  .  .  .  .  Name: "req3"
  1135  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
  1136  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
  1137  .  .  .  .  .  .  .  .  .  .  .  .  Name: "req3"
  1138  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 1129)
  1139  .  .  .  .  .  .  .  .  .  .  .  .  Data: 0
  1140  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
  1141  .  .  .  .  .  .  .  .  .  .  .  }
  1142  .  .  .  .  .  .  .  .  .  .  }
  1143  .  .  .  .  .  .  .  .  .  }
  1144  .  .  .  .  .  .  .  .  .  Type: *ast.SelectorExpr {
  1145  .  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
  1146  .  .  .  .  .  .  .  .  .  .  .  NamePos: foo:50:18
  1147  .  .  .  .  .  .  .  .  .  .  .  Name: "dto"
  1148  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
  1149  .  .  .  .  .  .  .  .  .  .  }
  1150  .  .  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
  1151  .  .  .  .  .  .  .  .  .  .  .  NamePos: foo:50:22
  1152  .  .  .  .  .  .  .  .  .  .  .  Name: "Good3Req"
  1153  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
  1154  .  .  .  .  .  .  .  .  .  .  }
  1155  .  .  .  .  .  .  .  .  .  }
  1156  .  .  .  .  .  .  .  .  .  Values: nil
  1157  .  .  .  .  .  .  .  .  .  Comment: nil
  1158  .  .  .  .  .  .  .  .  }
  1159  .  .  .  .  .  .  .  }
  1160  .  .  .  .  .  .  .  Rparen: -
  1161  .  .  .  .  .  .  }
  1162  .  .  .  .  .  }
  1163  .  .  .  .  .  3: *ast.AssignStmt {
  1164  .  .  .  .  .  .  Lhs: []ast.Expr (len = 1) {
  1165  .  .  .  .  .  .  .  0: *ast.Ident {
  1166  .  .  .  .  .  .  .  .  NamePos: foo:51:9
  1167  .  .  .  .  .  .  .  .  Name: "req4"
  1168  .  .  .  .  .  .  .  .  Obj: *ast.Object {
  1169  .  .  .  .  .  .  .  .  .  Kind: var
  1170  .  .  .  .  .  .  .  .  .  Name: "req4"
  1171  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 1163)
  1172  .  .  .  .  .  .  .  .  .  Data: nil
  1173  .  .  .  .  .  .  .  .  .  Type: nil
  1174  .  .  .  .  .  .  .  .  }
  1175  .  .  .  .  .  .  .  }
  1176  .  .  .  .  .  .  }
  1177  .  .  .  .  .  .  TokPos: foo:51:14
  1178  .  .  .  .  .  .  Tok: :=
  1179  .  .  .  .  .  .  Rhs: []ast.Expr (len = 1) {
  1180  .  .  .  .  .  .  .  0: *ast.CompositeLit {
  1181  .  .  .  .  .  .  .  .  Type: *ast.SelectorExpr {
  1182  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
  1183  .  .  .  .  .  .  .  .  .  .  NamePos: foo:51:17
  1184  .  .  .  .  .  .  .  .  .  .  Name: "dto"
  1185  .  .  .  .  .  .  .  .  .  .  Obj: nil
  1186  .  .  .  .  .  .  .  .  .  }
  1187  .  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
  1188  .  .  .  .  .  .  .  .  .  .  NamePos: foo:51:21
  1189  .  .  .  .  .  .  .  .  .  .  Name: "Good3Req"
  1190  .  .  .  .  .  .  .  .  .  .  Obj: nil
  1191  .  .  .  .  .  .  .  .  .  }
  1192  .  .  .  .  .  .  .  .  }
  1193  .  .  .  .  .  .  .  .  Lbrace: foo:51:29
  1194  .  .  .  .  .  .  .  .  Elts: nil
  1195  .  .  .  .  .  .  .  .  Rbrace: foo:51:30
  1196  .  .  .  .  .  .  .  .  Incomplete: false
  1197  .  .  .  .  .  .  .  }
  1198  .  .  .  .  .  .  }
  1199  .  .  .  .  .  }
*/

// ReqResInfo
type ReqResInfo struct {
	ModName   string
	ParamName string
}

// getStructByParamName 根据赋值变量名获取他的结构体
func getStructByParamName(paramName string, list []ast.Stmt) ReqResInfo {
	var output ReqResInfo
	for _, value := range list {
		fmt.Printf("value--------------->"+"%+v\n", value)
		switch stmtType := value.(type) {
		case *ast.DeclStmt:
			if stmtType.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Names[0].String() == paramName {
				switch specType := stmtType.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Type.(type) {
				// var req GoodCreateReq
				case *ast.Ident:
					output.ParamName = specType.Name
				// var req3 dto.Good3Req
				case *ast.SelectorExpr:
					output.ModName = specType.X.(*ast.Ident).String()
					output.ParamName = specType.Sel.String()
				}
			}
		case *ast.AssignStmt:
			if stmtType.Lhs[0].(*ast.Ident).String() == paramName {
				fmt.Printf("paramName--------------->"+"%+v\n", paramName)
				fmt.Printf(" stmtType.Rhs[0]--------------->"+"%+v\n", list)
				switch RhsType := stmtType.Rhs[0].(type) {
				case *ast.CompositeLit:
					switch specType := RhsType.Type.(type) {
					// req :=  GoodCreateReq
					case *ast.Ident:
						output.ParamName = specType.Name
					// req3 :=  dto.Good3Req
					case *ast.SelectorExpr:
						output.ModName = specType.X.(*ast.Ident).String()
						output.ParamName = specType.Sel.String()
					}
				case *ast.UnaryExpr:
					switch rhsTypeXType := RhsType.X.(type) {
					case *ast.CompositeLit:
						switch specType := rhsTypeXType.Type.(type) {
						// req :=  GoodCreateReq
						case *ast.Ident:
							output.ParamName = specType.Name
						// req3 :=  dto.Good3Req
						case *ast.SelectorExpr:
							output.ModName = specType.X.(*ast.Ident).String()
							output.ParamName = specType.Sel.String()
						}
					}
				}

			}

		}
	}
	return output
}
