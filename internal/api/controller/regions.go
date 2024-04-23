package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

func (c *Controller) GetProvidersRegions(ctx echo.Context) error {
	regions, err := c.service.ListProvidersRegions(ctx.Request().Context())
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, regions)
}

func (c *Controller) GetCategoriesByRegionID(ctx echo.Context) error {
	id := ctx.Param("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		idInt = 0
	}

	categories, err := c.service.ListCategoriesByRegionID(ctx.Request().Context(), int64(idInt))
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, categories)
}
