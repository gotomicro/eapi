package main

import (
	"os"

	"github.com/gotomicro/eapi"
	"github.com/gotomicro/eapi/plugins/echo"
	"github.com/gotomicro/eapi/plugins/gin"
)

func main() {
	eapi.NewEntrypoint(
		gin.NewPlugin(),
		echo.NewPlugin(),
	).Run(os.Args)
}
