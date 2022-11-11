package main

import (
	"os"

	analyzer "github.com/gotomicro/ego-gen-api"
	"github.com/gotomicro/ego-gen-api/plugins/gin"
)

func main() {
	analyzer.NewEntrypoint(
		gin.NewPlugin(),
	).Run(os.Args)
}
