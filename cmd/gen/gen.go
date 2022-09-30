package gen

import (
	"fmt"

	"ego-gen-api/cmd"
	"ego-gen-api/internal/parser"
	"github.com/spf13/cobra"
)

var CmdRun = &cobra.Command{
	Use:   "gen",
	Short: "生成前端代码",
	Long:  `启动`,
	Run:   CmdFunc,
}

var path string
var main string

func init() {
	CmdRun.InheritedFlags()
	CmdRun.PersistentFlags().StringVarP(&path, "path", "p", "", "指定路径")
	CmdRun.PersistentFlags().StringVarP(&main, "main", "m", "main.go", "指定main文件")
	cmd.RootCommand.AddCommand(CmdRun)
}

func CmdFunc(cmd *cobra.Command, args []string) {
	if path == "" {
		fmt.Println("路径不能为空")
		return
	}

	parser.AstParserBuild(parser.UserOption{
		RootMainGo: main,
		RootPath:   path,
	})
}
