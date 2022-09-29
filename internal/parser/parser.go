package parser

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/samber/lo"
)

type astParser struct {
	readContent string // 读取原文件数据
	userOption  UserOption
	importMap   map[string]struct{}
	routerFunc  *ast.FuncDecl
	routerParam string // router的变量名
	// 存储group信息
	groupMap map[string]GroupInfo
	// 存储url信息
	urlMap map[string]UrlInfo
}

type GroupInfo struct {
	ParamName string // 变量名
	Prefix    string // 前缀
}

type UrlInfo struct {
	FullPath   string
	Method     string
	Prefix     string
	ModuleName string
	FuncName   string
}

func AstParserBuild(userOption UserOption) (*astParser, error) {
	a := &astParser{
		readContent: "",
		userOption:  userOption,
		importMap:   map[string]struct{}{},
		routerParam: "",
		groupMap:    map[string]GroupInfo{},
		urlMap:      map[string]UrlInfo{},
	}
	err := a.initReadContent()
	if err != nil {
		return nil, err
	}
	a.parserStruct()
	return a, nil
}

func (a *astParser) initReadContent() error {
	if a.userOption.ScaffoldDSLContent == "" {
		if a.userOption.RootFile == "" {
			return fmt.Errorf("content and file is empty")
		}
		contentByte, err := ioutil.ReadFile(a.userOption.RootFile)
		if err != nil {
			panic("initReadContent: " + err.Error())
		}
		a.userOption.ScaffoldDSLContent = string(contentByte)
		if a.userOption.ScaffoldDSLContent == "" {
			return fmt.Errorf("内容不能为空")
		}
	}
	a.readContent = a.userOption.ScaffoldDSLContent
	return nil
}

/**
*ast.File {
     1  .  Doc: nil
     2  .  Package: foo:1:1
     3  .  Name: *ast.Ident {
     4  .  .  NamePos: foo:1:9
     5  .  .  Name: "testdata"
     6  .  .  Obj: nil
     7  .  }
     8  .  Decls: []ast.Decl (len = 2) {
     9  .  .  0: *ast.GenDecl {
    10  .  .  .  Doc: nil
    11  .  .  .  TokPos: foo:3:1
    12  .  .  .  Tok: import
    13  .  .  .  Lparen: foo:3:8
    14  .  .  .  Specs: []ast.Spec (len = 2) {
    15  .  .  .  .  0: *ast.ImportSpec {
    16  .  .  .  .  .  Doc: nil
    17  .  .  .  .  .  Name: nil
    18  .  .  .  .  .  Path: *ast.BasicLit {
    19  .  .  .  .  .  .  ValuePos: foo:4:2
    20  .  .  .  .  .  .  Kind: STRING
    21  .  .  .  .  .  .  Value: "\"net/http\""
    22  .  .  .  .  .  }
    23  .  .  .  .  .  Comment: nil
    24  .  .  .  .  .  EndPos: -
    25  .  .  .  .  }
    26  .  .  .  .  1: *ast.ImportSpec {
    27  .  .  .  .  .  Doc: nil
    28  .  .  .  .  .  Name: nil
    29  .  .  .  .  .  Path: *ast.BasicLit {
    30  .  .  .  .  .  .  ValuePos: foo:5:2
    31  .  .  .  .  .  .  Kind: STRING
    32  .  .  .  .  .  .  Value: "\"strings\""
    33  .  .  .  .  .  }
    34  .  .  .  .  .  Comment: nil
    35  .  .  .  .  .  EndPos: -
    36  .  .  .  .  }
    37  .  .  .  }
    38  .  .  .  Rparen: foo:8:1
    39  .  .  }
    40  .  .  1: *ast.FuncDecl {
    41  .  .  .  Doc: nil
    42  .  .  .  Recv: nil
    43  .  .  .  Name: *ast.Ident {
    44  .  .  .  .  NamePos: foo:12:6
    45  .  .  .  .  Name: "ServeHttp"
    46  .  .  .  .  Obj: *ast.Object {
    47  .  .  .  .  .  Kind: func
    48  .  .  .  .  .  Name: "ServeHttp"
    49  .  .  .  .  .  Decl: *(obj @ 40)
    50  .  .  .  .  .  Data: nil
    51  .  .  .  .  .  Type: nil
    52  .  .  .  .  }
    53  .  .  .  }
    54  .  .  .  Type: *ast.FuncType {
    55  .  .  .  .  Func: foo:12:1
    56  .  .  .  .  Params: *ast.FieldList {
    57  .  .  .  .  .  Opening: foo:12:15
    58  .  .  .  .  .  List: nil
    59  .  .  .  .  .  Closing: foo:12:16
    60  .  .  .  .  }
    61  .  .  .  .  Results: *ast.FieldList {
    62  .  .  .  .  .  Opening: -
    63  .  .  .  .  .  List: []*ast.Field (len = 1) {
    64  .  .  .  .  .  .  0: *ast.Field {
    65  .  .  .  .  .  .  .  Doc: nil
    66  .  .  .  .  .  .  .  Names: nil
    67  .  .  .  .  .  .  .  Type: *ast.StarExpr {
    68  .  .  .  .  .  .  .  .  Star: foo:12:18
    69  .  .  .  .  .  .  .  .  X: *ast.SelectorExpr {
    70  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
    71  .  .  .  .  .  .  .  .  .  .  NamePos: foo:12:19
    72  .  .  .  .  .  .  .  .  .  .  Name: "egin"
    73  .  .  .  .  .  .  .  .  .  .  Obj: nil
    74  .  .  .  .  .  .  .  .  .  }
    75  .  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
    76  .  .  .  .  .  .  .  .  .  .  NamePos: foo:12:24
    77  .  .  .  .  .  .  .  .  .  .  Name: "Component"
    78  .  .  .  .  .  .  .  .  .  .  Obj: nil
    79  .  .  .  .  .  .  .  .  .  }
    80  .  .  .  .  .  .  .  .  }
    81  .  .  .  .  .  .  .  }
    82  .  .  .  .  .  .  .  Tag: nil
    83  .  .  .  .  .  .  .  Comment: nil
    84  .  .  .  .  .  .  }
    85  .  .  .  .  .  }
    86  .  .  .  .  .  Closing: -
    87  .  .  .  .  }
    88  .  .  .  }
    89  .  .  .  Body: *ast.BlockStmt {
    90  .  .  .  .  Lbrace: foo:12:34
    91  .  .  .  .  List: []ast.Stmt (len = 3) {
    92  .  .  .  .  .  0: *ast.AssignStmt {
    93  .  .  .  .  .  .  Lhs: []ast.Expr (len = 1) {
    94  .  .  .  .  .  .  .  0: *ast.Ident {
    95  .  .  .  .  .  .  .  .  NamePos: foo:13:2
    96  .  .  .  .  .  .  .  .  Name: "r"
    97  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    98  .  .  .  .  .  .  .  .  .  Kind: var
    99  .  .  .  .  .  .  .  .  .  Name: "r"
   100  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 92)
   101  .  .  .  .  .  .  .  .  .  Data: nil
   102  .  .  .  .  .  .  .  .  .  Type: nil
   103  .  .  .  .  .  .  .  .  }
   104  .  .  .  .  .  .  .  }
   105  .  .  .  .  .  .  }
   106  .  .  .  .  .  .  TokPos: foo:13:4
   107  .  .  .  .  .  .  Tok: :=
   108  .  .  .  .  .  .  Rhs: []ast.Expr (len = 1) {
   109  .  .  .  .  .  .  .  0: *ast.SelectorExpr {
   110  .  .  .  .  .  .  .  .  X: *ast.Ident {
   111  .  .  .  .  .  .  .  .  .  NamePos: foo:13:7
   112  .  .  .  .  .  .  .  .  .  Name: "invoker"
   113  .  .  .  .  .  .  .  .  .  Obj: nil
   114  .  .  .  .  .  .  .  .  }
   115  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
   116  .  .  .  .  .  .  .  .  .  NamePos: foo:13:15
   117  .  .  .  .  .  .  .  .  .  Name: "Gin"
   118  .  .  .  .  .  .  .  .  .  Obj: nil
   119  .  .  .  .  .  .  .  .  }
   120  .  .  .  .  .  .  .  }
   121  .  .  .  .  .  .  }
   122  .  .  .  .  .  }
   123  .  .  .  .  .  1: *ast.ExprStmt {
   124  .  .  .  .  .  .  X: *ast.CallExpr {
   125  .  .  .  .  .  .  .  Fun: *ast.SelectorExpr {
   126  .  .  .  .  .  .  .  .  X: *ast.Ident {
   127  .  .  .  .  .  .  .  .  .  NamePos: foo:16:2
   128  .  .  .  .  .  .  .  .  .  Name: "r"
   129  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 97)
   130  .  .  .  .  .  .  .  .  }
   131  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
   132  .  .  .  .  .  .  .  .  .  NamePos: foo:16:4
   133  .  .  .  .  .  .  .  .  .  Name: "POST"
   134  .  .  .  .  .  .  .  .  .  Obj: nil
   135  .  .  .  .  .  .  .  .  }
   136  .  .  .  .  .  .  .  }
   137  .  .  .  .  .  .  .  Lparen: foo:16:8
   138  .  .  .  .  .  .  .  Args: []ast.Expr (len = 2) {
   139  .  .  .  .  .  .  .  .  0: *ast.BasicLit {
   140  .  .  .  .  .  .  .  .  .  ValuePos: foo:16:9
   141  .  .  .  .  .  .  .  .  .  Kind: STRING
   142  .  .  .  .  .  .  .  .  .  Value: "\"/api/callbacks/wechat\""
   143  .  .  .  .  .  .  .  .  }
   144  .  .  .  .  .  .  .  .  1: *ast.CallExpr {
   145  .  .  .  .  .  .  .  .  .  Fun: *ast.Ident {
   146  .  .  .  .  .  .  .  .  .  .  NamePos: foo:16:34
   147  .  .  .  .  .  .  .  .  .  .  Name: "H"
   148  .  .  .  .  .  .  .  .  .  .  Obj: nil
   149  .  .  .  .  .  .  .  .  .  }
   150  .  .  .  .  .  .  .  .  .  Lparen: foo:16:35
   151  .  .  .  .  .  .  .  .  .  Args: []ast.Expr (len = 1) {
   152  .  .  .  .  .  .  .  .  .  .  0: *ast.SelectorExpr {
   153  .  .  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
   154  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: foo:16:36
   155  .  .  .  .  .  .  .  .  .  .  .  .  Name: "shop"
   156  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   157  .  .  .  .  .  .  .  .  .  .  .  }
   158  .  .  .  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
   159  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: foo:16:41
   160  .  .  .  .  .  .  .  .  .  .  .  .  Name: "CallbackWechat"
   161  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   162  .  .  .  .  .  .  .  .  .  .  .  }
   163  .  .  .  .  .  .  .  .  .  .  }
   164  .  .  .  .  .  .  .  .  .  }
   165  .  .  .  .  .  .  .  .  .  Ellipsis: -
   166  .  .  .  .  .  .  .  .  .  Rparen: foo:16:55
   167  .  .  .  .  .  .  .  .  }
   168  .  .  .  .  .  .  .  }
   169  .  .  .  .  .  .  .  Ellipsis: -
   170  .  .  .  .  .  .  .  Rparen: foo:16:56
   171  .  .  .  .  .  .  }
   172  .  .  .  .  .  }
   173  .  .  .  .  .  2: *ast.ReturnStmt {
   174  .  .  .  .  .  .  Return: foo:18:2
   175  .  .  .  .  .  .  Results: []ast.Expr (len = 1) {
   176  .  .  .  .  .  .  .  0: *ast.Ident {
   177  .  .  .  .  .  .  .  .  NamePos: foo:18:9
   178  .  .  .  .  .  .  .  .  Name: "r"
   179  .  .  .  .  .  .  .  .  Obj: *(obj @ 97)
   180  .  .  .  .  .  .  .  }
   181  .  .  .  .  .  .  }
   182  .  .  .  .  .  }
   183  .  .  .  .  }
   184  .  .  .  .  Rbrace: foo:19:1
   185  .  .  .  }
   186  .  .  }
   187  .  }
   188  .  Scope: *ast.Scope {
   189  .  .  Outer: nil
   190  .  .  Objects: map[string]*ast.Object (len = 1) {
   191  .  .  .  "ServeHttp": *(obj @ 46)
   192  .  .  }
   193  .  }
   194  .  Imports: []*ast.ImportSpec (len = 2) {
   195  .  .  0: *(obj @ 15)
   196  .  .  1: *(obj @ 26)
   197  .  }
   198  .  Unresolved: []*ast.Ident (len = 4) {
   199  .  .  0: *(obj @ 70)
   200  .  .  1: *(obj @ 110)
   201  .  .  2: *(obj @ 145)
   202  .  .  3: *(obj @ 153)
   203  .  }
   204  .  Comments: []*ast.CommentGroup (len = 2) {
   205  .  .  0: *ast.CommentGroup {
   206  .  .  .  List: []*ast.Comment (len = 1) {
   207  .  .  .  .  0: *ast.Comment {
   208  .  .  .  .  .  Slash: foo:10:1
   209  .  .  .  .  .  Text: "//var H = bffcore.Handle"
   210  .  .  .  .  }
   211  .  .  .  }
   212  .  .  }
   213  .  .  1: *ast.CommentGroup {
   214  .  .  .  List: []*ast.Comment (len = 1) {
   215  .  .  .  .  0: *ast.Comment {
   216  .  .  .  .  .  Slash: foo:16:58
   217  .  .  .  .  .  Text: "// 微信通知回调接口"
   218  .  .  .  .  }
   219  .  .  .  }
   220  .  .  }
   221  .  }
   222  }
*/
func (p *astParser) parserStruct() error {
	fSet := token.NewFileSet()

	// strings.NewReader
	f, err := parser.ParseFile(fSet, "", strings.NewReader(p.readContent), parser.ParseComments)
	if err != nil {
		panic(err)
	}
	// Print the AST.
	var bf bytes.Buffer
	ast.Fprint(&bf, fSet, f, func(string, reflect.Value) bool {
		return true
	})

	commentMap := ast.NewCommentMap(fSet, f, f.Comments)
	fmt.Printf("commentMap--------------->"+"%+v\n", commentMap)
	f.Comments = commentMap.Filter(f).Comments()

	funcValue, orderNum := p.findEginRouter(f)
	p.findRouterParam(funcValue, orderNum)
	//p.findGroup()
	fmt.Printf("p--------------->"+"%+v\n", p.groupMap)
	for _, stmtValue := range funcValue.Body.List {
		p.dealGroup(stmtValue)
		p.findUrl(stmtValue)
		//stmt, flag := stmtValue.(*ast.ExprStmt)
		//if !flag {
		//	continue
		//}
		//stmtCall, flag := stmt.X.(*ast.CallExpr)
		//if !flag {
		//	continue
		//}
		//funcSelector, flag := stmtCall.Fun.(*ast.SelectorExpr)
		//if !flag {
		//	continue
		//}
		//routerIndent, flag := funcSelector.X.(*ast.Ident)
		//if !flag {
		//	continue
		//}
		//
		//// 找到变量，在找
		//if routerIndent.Name == p.routerParam {
		//	fmt.Println("111111")
		//}
		//.Fun.(*ast.SelectorExpr).Sel.Name = "POST"
		//stmt.X.(*ast.CallExpr).Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name = "H"
	}
	fmt.Printf("funcValue--------------->"+"%+v\n", funcValue)

	scope := f.Scope
	if scope == nil {
		return errors.New("struct nil")
	}
	//objects := scope.Objects
	//fmt.Printf("objects--------------->"+"%+v\n", objects["ServeHttp"])

	//structs := make([]*SpecType, 0)
	//for structName, obj := range objects {
	//	st, err := a.parseObject(structName, obj)
	//	if err != nil {
	//		return err
	//	}
	//	structs = append(structs, st)
	//}
	//sort.Slice(structs, func(i, j int) bool {
	//	return structs[i].Name < structs[j].Name
	//})
	//
	//resp := make([]SpecType, 0)
	//for _, item := range structs {
	//	resp = append(resp, *item)
	//}
	//a.modelArr = resp
	return nil
}

/**
如何确定一个ego的gin router
在返回值中，只要有一个参数是*egin.Component，就是我们的router函数
然后我们在从return中找到对应顺序的变量名
在然后通过解析这个函数里数据，如果是这个变量名，并且是函数调用方法，使用的Group、GET、POST等名称，那么就是我们的方法
再然后通过import找到对应的文件路径，解析他的dto
*/
func (p *astParser) findEginRouter(f *ast.File) (*ast.FuncDecl, int) {
	//  f.Decls
	for _, value := range f.Decls {
		// *ast.FuncDecl
		// 拿到函数
		valueData, flag := value.(*ast.FuncDecl)
		if !flag {
			continue
		}
		// HttpServer
		// valueData.Name
		// 这里放着参数、返回值
		// orderNum是第几个顺序
		// 因为是多个返回值，所以我们需要遍历*ast.Field，我们认为*egin.Component，就是我们的所有router项目
		for orderNum, fieldValue := range valueData.Type.Results.List {
			resultFieldPointer, flag := fieldValue.Type.(*ast.StarExpr)
			if !flag {
				continue
			}
			name := resultFieldPointer.X.(*ast.SelectorExpr).X.(*ast.Ident).Name
			if name == "egin" {
				// 这个才是函数router
				if resultFieldPointer.X.(*ast.SelectorExpr).Sel.Name == "Component" {
					p.routerFunc = valueData
					return valueData, orderNum
				}
			}
		}
	}
	panic("no ego component")
	return nil, 0
}

// 找到变量名
func (p *astParser) findRouterParam(decl *ast.FuncDecl, orderNum int) {
	for _, stmtValue := range decl.Body.List {
		// 找到返回值
		stmt, flag := stmtValue.(*ast.ReturnStmt)
		if !flag {
			continue
		}
		p.routerParam = stmt.Results[orderNum].(*ast.Ident).Name
		// 因为只有一个返回值直接return
		// 后续可以通过
		return
	}
	panic("no router param")
	return
}

// 找到group
func (p *astParser) findGroup() {
	for _, stmtValue := range p.routerFunc.Body.List {
		stmt, flag := stmtValue.(*ast.AssignStmt)
		if !flag {
			continue
		}

		// g := r.Group()
		if !(stmt.Tok.String() == ":=") {
			continue
		}

		assignValue := stmt.Rhs[0]
		stmtCall, flag := assignValue.(*ast.CallExpr)
		if !flag {
			continue
		}

		funcSelector, flag := stmtCall.Fun.(*ast.SelectorExpr)
		if !flag {
			continue
		}
		// r
		routerIndent, flag := funcSelector.X.(*ast.Ident)
		if !flag {
			continue
		}
		// 说明是router参数的名称，说明是分组
		if routerIndent.Name == p.routerParam && funcSelector.Sel.Name == "Group" {
			keyValue := stmt.Lhs[0].(*ast.Ident).Name
			p.groupMap[keyValue] = GroupInfo{
				keyValue,
				stmtCall.Args[0].(*ast.BasicLit).Value,
			}
		}
		return
	}
	return
}

/**
1. g := r.Group()
2. g = g.Group()
*/

func (p *astParser) dealGroup(stmtValue ast.Stmt) {
	stmt, flag := stmtValue.(*ast.AssignStmt)
	if !flag {
		return
	}

	// g := r.Group()
	if !(stmt.Tok.String() == ":=") && !(stmt.Tok.String() == "=") {
		return
	}

	assignValue := stmt.Rhs[0]
	stmtCall, flag := assignValue.(*ast.CallExpr)
	if !flag {
		return
	}

	funcSelector, flag := stmtCall.Fun.(*ast.SelectorExpr)
	if !flag {
		return
	}
	// r
	routerIndent, flag := funcSelector.X.(*ast.Ident)
	if !flag {
		return
	}

	// 去掉引号

	if stmt.Tok.String() == ":=" {
		// 说明是router参数的名称，说明是分组
		if routerIndent.Name == p.routerParam && funcSelector.Sel.Name == "Group" {
			keyValue := stmt.Lhs[0].(*ast.Ident).Name
			p.groupMap[keyValue] = GroupInfo{
				keyValue,
				strings.Trim(stmtCall.Args[0].(*ast.BasicLit).Value, `"`),
			}
			return
		}
	}

	if stmt.Tok.String() == "=" {
		// 说明是之前旧的router参数的名称，说明是分组
		groupInfo, flag := p.groupMap[routerIndent.Name]
		if flag {
			if funcSelector.Sel.Name == "Group" {
				keyValue := stmt.Lhs[0].(*ast.Ident).Name
				p.groupMap[keyValue] = GroupInfo{
					keyValue,
					groupInfo.Prefix + strings.Trim(stmtCall.Args[0].(*ast.BasicLit).Value, `"`),
				}
				return
			}
		}
	}
}

// 找到group
func (p *astParser) findUrl(stmtValue ast.Stmt) {
	stmt, flag := stmtValue.(*ast.ExprStmt)
	if !flag {
		return
	}
	callExpr, flag := stmt.X.(*ast.CallExpr)
	if !flag {
		return
	}

	funcSelector, flag := callExpr.Fun.(*ast.SelectorExpr)
	if !flag {
		return
	}
	// r
	routerIndent, flag := funcSelector.X.(*ast.Ident)
	if !flag {
		return
	}

	// 如果是根目录
	if routerIndent.Name == p.routerParam {
		p.getUrl(funcSelector.Sel.Name, "", callExpr)
		return
	}

	groupInfo, flag := p.groupMap[routerIndent.Name]
	if flag {
		p.getUrl(funcSelector.Sel.Name, groupInfo.Prefix, callExpr)
		return
	}
	return
}

func (p *astParser) getUrl(methodName string, prefix string, callExpr *ast.CallExpr) {
	if lo.Contains([]string{"POST", "GET", "DELETE", "PUT", "PATCH", "OPTION"}, methodName) {
		length := len(callExpr.Args)
		if length < 2 {
			panic("get url length < 2")
		}
		urlKey, flag := callExpr.Args[0].(*ast.BasicLit)
		if !flag {
			panic("get url type fail")
		}
		urlPath := strings.Trim(urlKey.Value, `"`)
		fullPath := prefix + urlPath
		funcInfo := getFuncType(callExpr.Args[length-1])
		p.urlMap[fullPath] = UrlInfo{
			Method:     methodName,
			Prefix:     prefix,
			FullPath:   fullPath,
			ModuleName: funcInfo.ModuleName,
			FuncName:   funcInfo.FuncName,
		}
		return
	}
}

type CallApiFunc struct {
	ModuleName string
	FuncName   string
}

func getFuncType(expr ast.Expr) CallApiFunc {
	switch v := expr.(type) {
	case *ast.SelectorExpr:
		x := v.X
		xIdent, ok := x.(*ast.Ident)
		if !ok {
			panic("get func type fail")
		}
		return CallApiFunc{
			ModuleName: xIdent.Name,
			FuncName:   v.Sel.Name,
		}
	case *ast.CallExpr:
		arg := v.Args[0]
		return getFuncType(arg)
	}
	panic("not exist type")
}
