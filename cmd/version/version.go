package version

import (
	"fmt"

	"ego-gen-api/cmd"

	"github.com/gotomicro/ego/core/eapp"
	"github.com/spf13/cobra"
)

const verboseVersionBanner string = `%s%s
%s
%s%s
├── Version   	    : %s 
├── BuildStatus     : %s 
└── BuildTime       : %s 
`

const shortVersionBanner = `
`

var CmdVersion = &cobra.Command{
	Use:   "version",
	Short: "打印当前版本信息",
	Long: `
打印当前程序版本信息
`,
	Run: versionCmd,
}

func init() {
	cmd.RootCommand.AddCommand(CmdVersion)
}

func versionCmd(cmd *cobra.Command, args []string) {
	coloredBanner := fmt.Sprintf(verboseVersionBanner, "\x1b[35m", "\x1b[1m",
		"\x1b[0m", "\x1b[32m", "\x1b[1m", eapp.AppVersion(), eapp.BuildStatus(), eapp.BuildTime())
	println(coloredBanner)
	return
}

// ShowShortVersionBanner prints the short version banner.
func ShowShortVersionBanner() {
	println(shortVersionBanner)
}
