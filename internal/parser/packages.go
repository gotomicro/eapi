package parser

import (
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
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
	// testdata/bff/pkg/shop/shop.go
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
		return nil, fmt.Errorf("not packages found" + pkgPath)
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
// 根据pkgName通过import path中，找到对应的ast file文件
// 例如
// import v1 "xxx.com/pb/common/v1"
// pkg 为v1
// 返回 xxx.com/pb/common/v1
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

// 根据v1.FileInfo，寻找
// 先判断是不是基础烈性
// 判断是不是有
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
		//isAliasPkgName := func(file *ast.File, pkgName string) bool {
		//	if file != nil && file.Imports != nil {
		//		for _, pkg := range file.Imports {
		//			if pkg.Name != nil {
		//				fmt.Printf("pkg.Name.Name--------------->"+"%+v\n", pkg.Name.Name)
		//				//return true
		//			}
		//			// pkg.Name 说明是alias 名字，例如 sts "google.golang.org/grpc/status" , 那么 sts 就是 alias
		//			if pkg.Name != nil && pkg.Name.Name == pkgName {
		//				return true
		//			}
		//		}
		//	}
		//
		//	return false
		//}

		if !IsAliasPkgName(file, parts[0]) {
			typeDef, ok := pkgDefs.uniqueDefinitions[typeName]
			if ok {
				return typeDef
			}
		}
		fmt.Printf("parts[0]--------------->"+"%+v\n", parts[0])
		/*
			parts[0]--------------->v1
			pkgPath--------------->xxx.com/pb/common/v1
		*/
		pkgPath := pkgDefs.findPackagePathFromImports(parts[0], file, false)
		//fmt.Printf("pkgPath--------------->"+"%+v\n", pkgPath)
		//fmt.Printf("file.Name.Name--------------->"+"%+v\n", file.Name.Name)
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

		fmt.Printf("parts[1]--------------->"+"%+v\n", parts[1])
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

func IsAliasPkgName(file *ast.File, pkgName string) bool {
	if file != nil && file.Imports != nil {
		for _, pkg := range file.Imports {
			if pkg.Name != nil {
				fmt.Printf("pkg.Name.Name--------------->"+"%+v\n", pkg.Name.Name)
				//return true
			}
			// pkg.Name 说明是alias 名字，例如 sts "google.golang.org/grpc/status" , 那么 sts 就是 alias
			if pkg.Name != nil && pkg.Name.Name == pkgName {
				return true
			}
		}
	}

	return false
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

// ReqResInfo 获取req，res信息
// ParamName就是结构体名称
// File就是结构体当前所在的file
// 尼玛会出现没名字情况。。要重新定义一个字段
type ReqResInfo struct {
	StructName    string
	IsArray       bool
	StructLocFile *ast.File // 结构体所在的文件路径
	FieldName     string
	FieldValue    string
	FieldInfo     string
	Type          int // 1 Func(), 2 xxx.FuncName(), 3 xxx.xxx.Func()
}

// 根据参数找到他对应的结构体
/*
	var param Struct
	var param dto.Struct
	var param *Struct
	var param *dto.Struct
	param := Struct{}
	param := &Struct{}
	param := dto.Struct{}
	param := &dto.Struct{}
	info = goodv1.GoodSku{}
	info ,err = invoker.Good.GoodSku(ctx, &goodv1.GoodSkuReq{})
	c.JSONOK(info.SpecList)
	c.JSONOK(info.GetSpecList) // 暂时不支持
*/

type ParamInfo struct {
	SelectorName string
	Type         int // 0 没有引用， 1 有引用， 2 是函数，不支持
	SecondName   string
}

// getStructByReferName 根据赋值变量名获取他的结构体
func (p *astParser) getStructByReferName(paramInfo ParamInfo, currentFile *ast.File, list []ast.Stmt) ReqResInfo {
	var output ReqResInfo
	for _, value := range list {
		switch stmtType := value.(type) {
		case *ast.DeclStmt:
			if stmtType.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Names[0].String() == paramInfo.SelectorName {
				switch specType := stmtType.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Type.(type) {
				// var req GoodCreateReq
				case *ast.Ident:
					output.StructName = specType.Name
					output.StructLocFile = currentFile
				// var req3 dto.Good3Req
				case *ast.SelectorExpr:
					modName := specType.X.(*ast.Ident).String()
					structName := specType.Sel.String()
					importMapInfo := p.getImport(currentFile)
					importInfo, flag := importMapInfo[modName]
					if !flag {
						panic("not find import info, modName: " + modName)
					}
					structLocFile, err := p.getLocFile(importInfo.PackagePath, modName, structName)
					if err != nil {
						panic("getLocFile" + err.Error())
					}
					output.StructName = specType.Sel.String()
					output.StructLocFile = structLocFile
				}
			}
			// list := make([]slice)
		case *ast.AssignStmt:
			// 没有prefix
			switch lhsType := stmtType.Lhs[0].(type) {
			case *ast.Ident:
				//identValue, flag := stmtType.Lhs[0].(*ast.Ident)
				if lhsType.String() == paramInfo.SelectorName {
					switch RhsType := stmtType.Rhs[0].(type) {
					case *ast.CompositeLit:
						switch specType := RhsType.Type.(type) {
						// req :=  GoodCreateReq
						case *ast.Ident:
							output.StructName = specType.Name
							output.StructLocFile = currentFile
						// req3 :=  dto.Good3Req
						case *ast.SelectorExpr:
							modName := specType.X.(*ast.Ident).String()
							structName := specType.Sel.String()
							importMapInfo := p.getImport(currentFile)
							importInfo, flag := importMapInfo[modName]
							if !flag {
								panic("not find import info, modName: " + modName)
							}

							structLocFile, err := p.getLocFile(importInfo.PackagePath, modName, structName)
							if err != nil {
								panic("getLocFile" + err.Error())
							}
							output.StructLocFile = structLocFile
							output.StructName = specType.Sel.String()
						// info := []Teacher{}
						case *ast.ArrayType:
							switch specType := specType.Elt.(type) {
							case *ast.Ident:
								output.StructName = specType.Name
								output.IsArray = true
								output.StructLocFile = currentFile
							case *ast.SelectorExpr:
								modName := specType.X.(*ast.Ident).String()
								structName := specType.Sel.String()
								importMapInfo := p.getImport(currentFile)
								importInfo, flag := importMapInfo[modName]
								if !flag {
									panic("not find import info, modName: " + modName)
								}

								structLocFile, err := p.getLocFile(importInfo.PackagePath, modName, structName)
								if err != nil {
									panic("getLocFile" + err.Error())
								}
								output.StructLocFile = structLocFile
								output.IsArray = true
								output.StructName = specType.Sel.String()
							}
						}
					case *ast.UnaryExpr:
						switch rhsTypeXType := RhsType.X.(type) {
						case *ast.CompositeLit:
							switch specType := rhsTypeXType.Type.(type) {
							// req :=  &GoodCreateReq{}
							case *ast.Ident:
								output.StructName = specType.Name
								output.StructLocFile = currentFile
							// req3 :=  &dto.Good3Req{}
							case *ast.SelectorExpr:
								modName := specType.X.(*ast.Ident).String()
								structName := specType.Sel.String()
								importMapInfo := p.getImport(currentFile)
								importInfo, flag := importMapInfo[modName]
								if !flag {
									panic("not find import info, modName: " + modName)
								}

								structLocFile, err := p.getLocFile(importInfo.PackagePath, modName, structName)
								if err != nil {
									panic("getLocFile" + err.Error())
								}
								output.StructLocFile = structLocFile
								output.StructName = specType.Sel.String()
							}
						}
					// list := make([]slice,0)
					case *ast.CallExpr:
						//var flag bool
						switch funType := RhsType.Fun.(type) {
						case *ast.Ident:
							// 函数名可能是关键字make
							// info := make([]slice,0)
							if funType.String() == "make" {
								// 获取第一个参数
								switch argType := RhsType.Args[0].(type) {
								case *ast.ArrayType:
									fieldType := getFieldTypeX(argType.Elt)
									output.StructName = p.getTrueTypeName(fieldType, currentFile)
									output.StructLocFile = currentFile
									output.IsArray = true
								default:
									panic("no support make type")
								}
							} else {
								// 当前file
								// info,err := XXXX()
								//output.ModName = ""
								//output.FuncName = funType.String()
								output, _ = p.getModNameAndParamNameByCurrentFile(funType.String(), paramInfo.SecondName, currentFile)
							}

						case *ast.SelectorExpr:
							switch xType := funType.X.(type) {
							case *ast.Ident:
								// info,err := invoker.XXXX()
								//output.ModName = xType.String()
								//output.FuncName = funType.Sel.String()
								funcName := funType.Sel.String()

								importMapInfo := p.getImport(currentFile)
								importInfo, flag := importMapInfo[xType.String()]
								if !flag {
									panic("not find import info, modName: " + xType.String())
								}
								p.packages.findFileByRangePackages(importInfo.PackagePath, func(filename string, file *ast.File) (bool, error) {
									output, flag = p.getModNameAndParamNameByCurrentFile(funcName, paramInfo.SecondName, currentFile)
									if flag {
										return false, nil
									}
									return true, nil
								})

							case *ast.SelectorExpr:
								// info,err := invoker.Grpc.XXXX()
								//output.ModName = xType.X.(*ast.Ident).String()
								paramName := xType.Sel.String()
								//output.FuncName = funType.Sel.String()
								funcName := funType.Sel.String()
								importMapInfo := p.getImport(currentFile)
								importInfo, flag := importMapInfo[xType.X.(*ast.Ident).String()]
								if !flag {
									panic("not find import info, modName: " + xType.X.(*ast.Ident).String())
								}

								p.packages.findFileByRangePackages(importInfo.PackagePath, func(filename string, file *ast.File) (bool, error) {
									// 先找到有定义的
									for _, declValue := range file.Decls {
										// 找到函数
										declValue, flag := declValue.(*ast.GenDecl)
										if !flag {
											continue
										}
										for _, specValue := range declValue.Specs {
											assertValueSpec, flag := specValue.(*ast.ValueSpec)
											if !flag {
												continue
											}
											if assertValueSpec.Names[0].String() != paramName {
												continue
											}
											switch specType := assertValueSpec.Type.(type) {
											// var grpc GoodClient
											case *ast.Ident:
												output, _ = p.getModNameAndParamNameByCurrentFile(funcName, paramInfo.SecondName, file)
												return false, nil
											// var grpc shopv1.GoodClient
											case *ast.SelectorExpr:
												importParamMapInfo := p.getImport(file)
												importParamInfo, flag := importParamMapInfo[specType.X.(*ast.Ident).String()]
												if !flag {
													panic("not find import info, modName: " + specType.X.(*ast.Ident).String())
												}
												fmt.Printf("importParamInfo.PackagePath--------------->"+"%+v\n", importParamInfo.PackagePath)
												p.packages.findFileByRangePackages(importParamInfo.PackagePath, func(filename string, file *ast.File) (bool, error) {
													fmt.Printf("filename--------------->"+"%+v\n", filename)
													fmt.Printf("funcName--------------->"+"%+v\n", funcName)
													output, flag = p.getModNameAndParamNameByCurrentFile(funcName, paramInfo.SecondName, file)
													if flag {
														fmt.Printf("output--------------->"+"%+v\n", output)
														return false, nil
													}
													return true, nil
												})
												return false, nil
											}
										}

									}
									return true, nil
								})

							}

						}
					}

				}
			}
			// 有prefix
			// c.JSONOK(res.List)

		}
	}
	return output
}

// 根据package name +struct name获取在哪个文件
// todo 这个方法跟 FindTypeSpec 是差不多的。这边还少了alias玩法
func (p *astParser) getLocFile(packagePath string, packageName string, structName string) (structLocFile *ast.File, err error) {
	// 遍历去找这个type类型
	return p.packages.findFileByRangePackages(packagePath, func(filename string, file *ast.File) (bool, error) {
		// 循环遍历
		var isContinueForeach = true
		ast.Inspect(file, func(n ast.Node) bool {
			// 获取package name
			if file.Name.String() == packageName {
				switch nn := n.(type) {
				case *ast.GenDecl:
					// 定义的地方
					if nn.Tok == token.TYPE {
						info, flag := nn.Specs[0].(*ast.TypeSpec)
						if !flag {
							return true
						}
						if info.Name.String() == structName {
							// 找到了，不需要再循环
							isContinueForeach = false
							return false
						}
					}
				}
			}
			return true
		})

		return isContinueForeach, nil
	})
}

// 根据函数名，在当前文件获取modName，paramName
// 如果有fieldname，说明取的struct name中的一个字段的结构体信息
// 目前返回值必须是第一个
func (p *astParser) getModNameAndParamNameByCurrentFile(funcName string, fieldName string, currentFile *ast.File) (output ReqResInfo, flag bool) {
	//  currentFile 有pb.go
	if strings.HasSuffix(p.packages.files[currentFile].Path, "pb.go") {
		ast.Inspect(currentFile, func(node ast.Node) bool {
			switch nn := node.(type) {
			case *ast.InterfaceType:
				for _, value := range nn.Methods.List {
					if value.Names[0].String() == funcName {
						ast.Inspect(currentFile, func(node ast.Node) bool {
							switch nn := node.(type) {
							case *ast.FuncDecl:
								if nn.Name.String() == funcName {
									if nn.Type.Results == nil {
										return true
									}
									fmt.Printf("funcName--------------->"+"%+v\n", funcName)
									// todo 目前只支持第一个
									rst := nn.Type.Results.List[0]
									spew.Dump(nn)
									// 找到函数后，第一个变量返回，但是这个变量不是在这个文件里面，所以不能直接用这个currentFile，我们需要先定位到这个变量在哪个文件里
									fmt.Printf("rst--------------->\n")
									switch rstType := rst.Type.(type) {
									case *ast.SelectorExpr:
										flag = true
										structName := rstType.Sel.String()
										packageDir := p.packages.files[currentFile].PackagePath
										structLocFile, err := p.getLocFile(packageDir, p.packages.files[currentFile].File.Name.String(), structName)
										if err != nil {
											panic("getLocFile" + err.Error())
										}
										if fieldName == "" {
											output.StructLocFile = currentFile
											output.StructName = rstType.Sel.String()
										} else {
											typeSpecDef := p.packages.FindTypeSpec(structName, currentFile, true)
											if typeSpecDef == nil {
												panic("not found typedef")
											}
											structInfo := typeSpecDef.TypeSpec.Type.(*ast.StructType)
											for _, field := range structInfo.Fields.List {
												if field.Names[0].String() != fieldName {
													continue
												}
												typeName, err := getFieldType(field.Type)
												if err != nil {
													panic("not support type")
												}
												output.StructLocFile = structLocFile
												output.StructName = typeName
											}
										}
									case *ast.Ident:
										flag = true
										structName := rstType.String()
										packageDir := p.packages.files[currentFile].PackagePath
										structLocFile, err := p.getLocFile(packageDir, p.packages.files[currentFile].File.Name.String(), structName)
										if err != nil {
											panic("getLocFile" + err.Error())
										}
										if fieldName == "" {
											output.StructLocFile = structLocFile
											output.StructName = structName
										} else {
											typeSpecDef := p.packages.FindTypeSpec(structName, currentFile, true)
											if typeSpecDef == nil {
												panic("not found typedef")
											}
											structInfo := typeSpecDef.TypeSpec.Type.(*ast.StructType)
											for _, field := range structInfo.Fields.List {
												if field.Names[0].String() != fieldName {
													continue
												}
												typeName, err := getFieldType(field.Type)
												if err != nil {
													panic("not support type")
												}
												output.StructLocFile = structLocFile
												output.StructName = typeName
											}
										}
									// c.JSONOK(res)
									// c.JSONOK(res.List)
									case *ast.StarExpr:
										flag = true
										switch XType := rstType.X.(type) {
										case *ast.Ident:
											structName := XType.String()
											packageDir := p.packages.files[currentFile].PackagePath
											structLocFile, err := p.getLocFile(packageDir, p.packages.files[currentFile].File.Name.String(), structName)
											if err != nil {
												panic("getLocFile" + err.Error())
											}
											output.StructLocFile = structLocFile
											output.StructName = structName
											output.FieldName = fieldName
											if fieldName != "" {
												output.StructLocFile = structLocFile
												typeSpecDef := p.packages.FindTypeSpec(structName, currentFile, true)
												if typeSpecDef == nil {
													panic("not found typedef")
												}
												structInfo := typeSpecDef.TypeSpec.Type.(*ast.StructType)
												for _, field := range structInfo.Fields.List {
													if field.Names[0].String() != fieldName {
														continue
													}
													fieldValue := ""
													typeName, err := getFieldType(field.Type)
													if err != nil {
														// 数组类型
														// resp, err := invoker.xxx.List()
														// resp.List是数组
														arrType, flag := field.Type.(*ast.ArrayType)
														if !flag {
															panic("not support type")
														}
														fieldType := getFieldTypeX(arrType.Elt)
														output.FieldInfo = p.getTrueTypeName(fieldType, structLocFile)
														output.IsArray = true

														//
														//switch eltType := arrType.Elt.(type) {
														//case *ast.StarExpr:
														//	fieldType := getFieldTypeX(eltType.X)
														//	output.FieldInfo = p.getTrueTypeName(fieldType, structLocFile)
														//	//switch eltTypeX := eltType.X.(type) {
														//	//case *ast.Ident:
														//	//	output.FieldValue = eltTypeX.String()
														//	//	fieldType = eltTypeX.String()
														//	//case *ast.SelectorExpr:
														//	//	output.FieldValue = eltTypeX.Sel.String()
														//	//	fieldType = eltTypeX.X.(*ast.Ident).String() + "." + eltTypeX.Sel.String()
														//	//default:
														//	//	panic("not support type elt type star expr")
														//	//}
														//case *ast.Ident:
														//	output.FieldValue = eltType.String()
														//	fieldType := eltType.String()
														//	output.FieldInfo = p.getTrueTypeName(fieldType, structLocFile)
														//}
													} else {
														fieldValue = p.getTrueTypeName(typeName, structLocFile)
														fmt.Printf("typeName--------------->"+"%+v\n", typeName)
														output.FieldValue = typeName
														output.FieldInfo = fieldValue
													}
												}
											}
											// todo
										case *ast.SelectorExpr:
											fieldType := getFieldTypeX(XType)
											typeSpecDef := p.packages.FindTypeSpec(fieldType, currentFile, true)
											if typeSpecDef == nil {
												panic("not found typedef")
											}
											//output.StructLocFile = typeSpecDef.File
											//output.StructName = typeSpecDef.Name()
											//fmt.Printf("typeSpecDef--------------->"+"%+v\n", typeSpecDef.File)
										}

									}
									return false
								}

							}
							return true
						})
						return false
					}
				}
			}
			return true
		})
	} else {
		ast.Inspect(currentFile, func(node ast.Node) bool {
			switch nn := node.(type) {
			case *ast.FuncDecl:
				if nn.Name.String() == funcName {
					if nn.Type.Results == nil {
						return true
					}
					fmt.Printf("funcName--------------->"+"%+v\n", funcName)
					// todo 目前只支持第一个
					rst := nn.Type.Results.List[0]
					spew.Dump(nn)
					// 找到函数后，第一个变量返回，但是这个变量不是在这个文件里面，所以不能直接用这个currentFile，我们需要先定位到这个变量在哪个文件里
					fmt.Printf("rst--------------->\n")
					switch rstType := rst.Type.(type) {
					case *ast.SelectorExpr:
						flag = true
						structName := rstType.Sel.String()
						packageDir := p.packages.files[currentFile].PackagePath
						structLocFile, err := p.getLocFile(packageDir, p.packages.files[currentFile].File.Name.String(), structName)
						if err != nil {
							panic("getLocFile" + err.Error())
						}
						if fieldName == "" {
							output.StructLocFile = currentFile
							output.StructName = rstType.Sel.String()
						} else {
							typeSpecDef := p.packages.FindTypeSpec(structName, currentFile, true)
							if typeSpecDef == nil {
								panic("not found typedef")
							}
							structInfo := typeSpecDef.TypeSpec.Type.(*ast.StructType)
							for _, field := range structInfo.Fields.List {
								if field.Names[0].String() != fieldName {
									continue
								}
								typeName, err := getFieldType(field.Type)
								if err != nil {
									panic("not support type")
								}
								output.StructLocFile = structLocFile
								output.StructName = typeName
							}
						}
					case *ast.Ident:
						flag = true
						structName := rstType.String()
						packageDir := p.packages.files[currentFile].PackagePath
						structLocFile, err := p.getLocFile(packageDir, p.packages.files[currentFile].File.Name.String(), structName)
						if err != nil {
							panic("getLocFile" + err.Error())
						}
						if fieldName == "" {
							output.StructLocFile = structLocFile
							output.StructName = structName
						} else {
							typeSpecDef := p.packages.FindTypeSpec(structName, currentFile, true)
							if typeSpecDef == nil {
								panic("not found typedef")
							}
							structInfo := typeSpecDef.TypeSpec.Type.(*ast.StructType)
							for _, field := range structInfo.Fields.List {
								if field.Names[0].String() != fieldName {
									continue
								}
								typeName, err := getFieldType(field.Type)
								if err != nil {
									panic("not support type")
								}
								output.StructLocFile = structLocFile
								output.StructName = typeName
							}
						}
					// c.JSONOK(res)
					// c.JSONOK(res.List)
					case *ast.StarExpr:
						flag = true
						switch XType := rstType.X.(type) {
						case *ast.Ident:
							structName := XType.String()
							packageDir := p.packages.files[currentFile].PackagePath
							structLocFile, err := p.getLocFile(packageDir, p.packages.files[currentFile].File.Name.String(), structName)
							if err != nil {
								panic("getLocFile" + err.Error())
							}
							output.StructLocFile = structLocFile
							output.StructName = structName
							output.FieldName = fieldName
							if fieldName != "" {
								output.StructLocFile = structLocFile
								typeSpecDef := p.packages.FindTypeSpec(structName, currentFile, true)
								if typeSpecDef == nil {
									panic("not found typedef")
								}
								structInfo := typeSpecDef.TypeSpec.Type.(*ast.StructType)
								for _, field := range structInfo.Fields.List {
									if field.Names[0].String() != fieldName {
										continue
									}
									fieldValue := ""
									typeName, err := getFieldType(field.Type)
									if err != nil {
										// 数组类型
										// resp, err := invoker.xxx.List()
										// resp.List是数组
										arrType, flag := field.Type.(*ast.ArrayType)
										if !flag {
											panic("not support type")
										}
										fieldType := getFieldTypeX(arrType.Elt)
										output.FieldInfo = p.getTrueTypeName(fieldType, structLocFile)
										output.IsArray = true

										//
										//switch eltType := arrType.Elt.(type) {
										//case *ast.StarExpr:
										//	fieldType := getFieldTypeX(eltType.X)
										//	output.FieldInfo = p.getTrueTypeName(fieldType, structLocFile)
										//	//switch eltTypeX := eltType.X.(type) {
										//	//case *ast.Ident:
										//	//	output.FieldValue = eltTypeX.String()
										//	//	fieldType = eltTypeX.String()
										//	//case *ast.SelectorExpr:
										//	//	output.FieldValue = eltTypeX.Sel.String()
										//	//	fieldType = eltTypeX.X.(*ast.Ident).String() + "." + eltTypeX.Sel.String()
										//	//default:
										//	//	panic("not support type elt type star expr")
										//	//}
										//case *ast.Ident:
										//	output.FieldValue = eltType.String()
										//	fieldType := eltType.String()
										//	output.FieldInfo = p.getTrueTypeName(fieldType, structLocFile)
										//}
									} else {
										fieldValue = p.getTrueTypeName(typeName, structLocFile)
										fmt.Printf("typeName--------------->"+"%+v\n", typeName)
										output.FieldValue = typeName
										output.FieldInfo = fieldValue
									}
								}
							}
							// todo
						case *ast.SelectorExpr:
							fieldType := getFieldTypeX(XType)
							typeSpecDef := p.packages.FindTypeSpec(fieldType, currentFile, true)
							if typeSpecDef == nil {
								panic("not found typedef")
							}
							//output.StructLocFile = typeSpecDef.File
							//output.StructName = typeSpecDef.Name()
							//fmt.Printf("typeSpecDef--------------->"+"%+v\n", typeSpecDef.File)
						}

					}
					return false
				}

			}
			return true
		})
	}
	return
}

// 获取真正的类型名称
// @structLocFile 类型所在的文件
// 例如：
/*
	import v1 "git.xxx.com/pb/common/v1"
	typeName = v1.StructName
*/
func (p *astParser) getTrueTypeName(typeName string, structLocFile *ast.File) (trueType string) {
	typeArr := strings.Split(typeName, ".")
	if len(typeArr) > 1 {
		if IsAliasPkgName(structLocFile, typeArr[0]) {
			pkgPath := p.packages.findPackagePathFromImports(typeArr[0], structLocFile, false)
			specDef := p.packages.findTypeSpec(pkgPath, typeArr[1])
			trueType = specDef.File.Name.String() + "." + typeArr[1]
		}
	} else {
		if !IsGolangPrimitiveType(typeName) {
			trueType = structLocFile.Name.String() + "." + typeName
		} else {
			trueType = typeName
		}
	}
	return
}
