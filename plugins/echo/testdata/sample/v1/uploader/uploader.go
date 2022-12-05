package uploader

import (
	"net/http"
	"sample/model"

	"github.com/labstack/echo/v4"
)

// UploadFile
// @tags Uploader
// @summary Upload File
func UploadFile(c echo.Context) error {
	_, _ = c.FormFile("file")
	// If set true, this file would be saved to lib
	_ = c.FormValue("saveToLib")

	c.JSON(http.StatusOK, model.UploadFileRes{})
	return nil
}
