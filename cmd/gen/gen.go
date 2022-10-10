package gen

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ego-gen-api/cmd"
	"ego-gen-api/internal/parser"
	"ego-gen-api/internal/pongo2"
	"ego-gen-api/internal/pongo2render"
	"ego-gen-api/internal/utils"

	"github.com/go-openapi/spec"
	"github.com/spf13/cobra"
)

var CmdRun = &cobra.Command{
	Use:   "gen",
	Short: "生成前端代码",
	Long:  `启动`,
	Run:   CmdFunc,
}

var path string
var tmplPath string
var main string
var flushSuffix string
var dependencies string
var resFuncs string

func init() {
	CmdRun.InheritedFlags()
	CmdRun.PersistentFlags().StringVarP(&path, "path", "p", "", "指定路径")
	CmdRun.PersistentFlags().StringVarP(&tmplPath, "tmpl", "t", "", "指定路径")
	CmdRun.PersistentFlags().StringVarP(&dependencies, "dependencies", "d", "", "指定依赖路径")
	CmdRun.PersistentFlags().StringVarP(&flushSuffix, "flushSuffix", "f", ".gen.ts", "指定路径")
	CmdRun.PersistentFlags().StringVarP(&resFuncs, "resFuncs", "r", "JSONOK", "指定路径")
	CmdRun.PersistentFlags().StringVarP(&main, "main", "m", "main.go", "指定main文件")
	cmd.RootCommand.AddCommand(CmdRun)
}

func CmdFunc(cmd *cobra.Command, args []string) {
	if path == "" {
		fmt.Println("路径不能为空")
		return
	}

	p, err := parser.AstParserBuild(parser.UserOption{
		RootMainGo:   main,
		RootPath:     path,
		Dependencies: dependencies,
		ResFuncs:     resFuncs,
	})
	if err != nil {
		fmt.Println("parser fail, err: " + err.Error())
		return
	}

	// 生成 OpenAPI 文档
	generateOpenAPIDoc(p.GetData(), p.GetDefinitions())

	// 获取目录
	render := pongo2render.NewRender(filepath.Dir(tmplPath))
	err = Exec(render, tmplPath, p.GetData(), p.GetDefinitions())
	if err != nil {
		panic(err)
	}
	fmt.Println("finish")
}

func generateOpenAPIDoc(data []parser.UrlInfo, definitions spec.Definitions) {
	model := spec.Swagger{}
	model.Swagger = "2.0"

	model.Definitions = definitions
	paths := &spec.Paths{
		Paths: make(map[string]spec.PathItem),
	}
	for _, urlInfo := range data {
		path, ok := paths.Paths[urlInfo.FullPath]
		if !ok {
			path = spec.PathItem{}
		}

		setPathOperation(&path, urlInfo)
		paths.Paths[urlInfo.FullPath] = path
	}

	model.Paths = paths
	docContent, _ := json.MarshalIndent(model, "", "  ")
	filePath := filepath.Join(tmplPath, "dist", "swagger.json")
	err := ioutil.WriteFile(filePath, docContent, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func setPathOperation(path *spec.PathItem, urlInfo parser.UrlInfo) {
	operation := urlInfo.GetOperationSpec()

	switch strings.ToUpper(urlInfo.Method) {
	case "GET":
		path.Get = operation
	case "PUT":
		path.Put = operation
	case "POST":
		path.Post = operation
	case "DELETE":
		path.Delete = operation
	case "OPTIONS":
		path.Options = operation
	case "HEAD":
		path.Head = operation
	case "PATCH":
		path.Patch = operation
	}
}

func Exec(render *pongo2render.Render, dirPth string, data []parser.UrlInfo, definitions spec.Definitions) error {
	var (
		err error
	)
	files, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return err
	}

	for _, fi := range files {
		if !strings.HasSuffix(fi.Name(), ".tmpl") {
			continue
		}
		err = ExecOne(render, dirPth, fi.Name(), data, definitions)
		if err != nil {
			return err
		}
	}
	return nil
}

func ExecOne(render *pongo2render.Render, dirPath string, tmplName string, data []parser.UrlInfo, definitions spec.Definitions) error {
	var (
		buf string
		err error
	)
	flushFile := dirPath + "/dist/" + strings.TrimRight(filepath.Base(tmplName), filepath.Ext(filepath.Base(tmplName))) + flushSuffix
	ctx := make(pongo2.Context)
	ctx["data"] = data
	ctx["definitions"] = definitions
	buf, err = render.Template(filepath.Base(tmplName)).Execute(ctx)
	if err != nil {
		return fmt.Errorf("Could not create the %s render tmpl , err: %w", tmplName, err)
	}

	output := []byte(buf)
	err = write(flushFile, output)
	if err != nil {
		return fmt.Errorf("创建文件失败, err: %w", err)
	}
	return nil
}

// write to file
func write(filename string, buf []byte) (err error) {

	filePath := filepath.Dir(filename)
	err = createPath(filePath)
	if err != nil {
		err = errors.New("write create path " + err.Error())
		return
	}

	filePathBak := filePath + "/bak"
	err = createPath(filePathBak)
	if err != nil {
		err = errors.New("write create path bak " + err.Error())
		return
	}

	if utils.IsExist(filename) {
		bakName := fmt.Sprintf("%s/%s.%s.bak", filePathBak, filepath.Base(filename), time.Now().Format("2006.01.02.15.04.05"))
		if err := os.Rename(filename, bakName); err != nil {
			err = errors.New("file is bak error, path is " + bakName)
			return err
		}
	}

	file, err := os.Create(filename)
	defer func() {
		err = file.Close()
		if err != nil {
			panic(err)
		}
	}()
	if err != nil {
		err = errors.New("write create file " + err.Error())
		return
	}

	err = ioutil.WriteFile(filename, buf, 0644)
	if err != nil {
		err = errors.New("write write file " + err.Error())
		return
	}
	return
}

// createPath 调用os.MkdirAll递归创建文件夹
func createPath(filePath string) error {
	if !utils.IsExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		return err
	}
	return nil
}
