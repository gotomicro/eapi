package goods

import (
	"context"
	"net/http"

	"sample/model"

	"github.com/labstack/echo/v4"
)

// List
// @tags Goods
// @summary List Goods
func List(c echo.Context) error {
	var request model.ListGoodsRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, model.ListGoodsResponse{})
	return nil
}

// Create
// @tags Goods
// @summary Create Goods
func Create(c echo.Context) error {
	var request model.CreateGoodsRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, model.GoodsInfo{})
	return nil
}

// Update
// @tags Goods
// @summary Update Goods
func Update(c echo.Context) error {
	var request model.UpdateGoodsRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, updateGoods(model.GoodsInfo{}))
	return nil
}

func updateGoods[T any](val T) []T {
	return []T{val}
}

// Delete
// @tags Goods
// @summary Delete Goods
func Delete(c echo.Context) error {
	goodsId := c.Param("id")
	_ = goodsId

	res := deleteGoods(c.Request().Context(), goodsId)
	c.JSON(http.StatusOK, res)
	return nil
}

func deleteGoods(context context.Context, id string) model.GenericTypeResponse[string] {
	return model.GenericTypeResponse[string]{
		Data:     "deleted",
		Metadata: map[string]any{},
	}
}
