package gen

import (
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

var name string

func init() {
	CmdRun.InheritedFlags()
	CmdRun.PersistentFlags().StringVarP(&name, "name", "n", "", "指定任务名")
	cmd.RootCommand.AddCommand(CmdRun)
}

func CmdFunc(cmd *cobra.Command, args []string) {
	parser.AstParserBuild(parser.UserOption{
		RootMainGo: "main.go",
		RootPath:   "testdata/bff",
	})
}
