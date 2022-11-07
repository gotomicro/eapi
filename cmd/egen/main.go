package main

import (
	"os"

	analyzer "github.com/gotomicro/ego-gen-api"
	"github.com/gotomicro/ego-gen-api/plugins/ego"
)

func main() {
	analyzer.NewEntrypoint(ego.NewPlugin()).Run(os.Args)
}
