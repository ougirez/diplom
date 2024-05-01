package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/ougirez/diplom/internal/pkg/logger"
	"net/http"
)

func (c *Controller) BackFillFGBUData(ctx echo.Context) error {
	regionItems, err := c.service.ParseAndSaveProviderItems(ctx.Request().Context(), "https://inform-raduga.ru/fgbu")
	if err != nil {
		logger.Error(ctx.Request().Context(), err)
		return err
	}

	return ctx.JSON(http.StatusOK, regionItems)
}
