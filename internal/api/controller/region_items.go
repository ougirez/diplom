package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func (c *Controller) GetRegionItems(ctx echo.Context) error {
	regionItems, err := c.service.ParseAndSaveProviderItems(ctx.Request().Context(), "https://inform-raduga.ru/fgbu")
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, regionItems)
}
