package main

import (
	"sample/v1/goods"
	"sample/v1/uploader"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	v1 := e.Group("/v1")

	// Uploader
	{
		v1.POST("/upload", uploader.UploadFile)
	}

	// Goods
	{
		v1.GET("/goods", goods.List)
		v1.POST("/goods", goods.Create)
		v1.PATCH("/goods", goods.Update)
		v1.DELETE("/goods/:id", goods.Delete)
	}

	e.Start(":8081")
}
