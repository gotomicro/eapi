package main

import (
	"bff/pkg/router"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
)

//  export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	if err := ego.New().Serve(router.ServeHttp()).Run(); err != nil {
		elog.Panic("startup", elog.FieldErr(err))
	}
}
