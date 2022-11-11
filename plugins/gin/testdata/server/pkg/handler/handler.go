package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CustomContext struct {
	*gin.Context
}

func (c *CustomContext) JSONOK(data interface{}) {
	c.JSON(http.StatusOK, data)
}

func Handler(handler func(c *CustomContext)) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(&CustomContext{Context: c})
	}
}
