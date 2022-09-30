package gen

import (
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
var dependences string

func init() {
	CmdRun.InheritedFlags()
	CmdRun.PersistentFlags().StringVarP(&path, "path", "p", "", "指定路径")
	CmdRun.PersistentFlags().StringVarP(&tmplPath, "tmpl", "t", "testdata/tmpls/api.tmpl", "指定路径")
	CmdRun.PersistentFlags().StringVarP(&dependences, "dependences", "d", "git.gocn.vip/of/pb", "指定依赖路径")
	CmdRun.PersistentFlags().StringVarP(&flushSuffix, "flushSuffix", "f", ".gen.ts", "指定路径")
	CmdRun.PersistentFlags().StringVarP(&main, "main", "m", "main.go", "指定main文件")
	cmd.RootCommand.AddCommand(CmdRun)
}

func CmdFunc(cmd *cobra.Command, args []string) {
	if path == "" {
		fmt.Println("路径不能为空")
		return
	}

	p, err := parser.AstParserBuild(parser.UserOption{
		RootMainGo:  main,
		RootPath:    path,
		Dependences: dependences,
	})
	if err != nil {
		fmt.Println("parser fail, err: " + err.Error())
		return
	}
	// 获取目录

	render := pongo2render.NewRender(filepath.Dir(tmplPath))
	err = Exec(render, tmplPath, p.GetData(), p.GetDefinitions())
	if err != nil {
		panic(err)
	}
	fmt.Println("finish")
}

func Exec(render *pongo2render.Render, tmplName string, data []parser.UrlInfo, definitions spec.Definitions) error {
	var (
		buf string
		err error
	)

	flushFile := filepath.Dir(tmplName) + "/dist/" + strings.TrimRight(filepath.Base(tmplName), filepath.Ext(filepath.Base(tmplName))) + flushSuffix
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
