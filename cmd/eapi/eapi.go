package main

import (
	"os"

	analyzer "github.com/gotomicro/eapi"
	"github.com/gotomicro/eapi/plugins/gin"
)

func main() {
	analyzer.NewEntrypoint(
		gin.NewPlugin(),
	).Run(os.Args)
}
