package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/ougirez/diplom/internal/pkg/store"
	"net/http"
	"strconv"
)

func (c *Controller) GetRegions(ctx echo.Context) error {
	regions, err := c.service.ListRegions(ctx.Request().Context())
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, regions)
}

func (c *Controller) GetCategoriesByRegionID(ctx echo.Context) error {
	id := ctx.Param("id")
	categoryName := ctx.QueryParams().Get("category_name")
	groupCategoryName := ctx.QueryParams().Get("group_category_name")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		idInt = 0
	}

	opts := store.ListCategoriesByRegionIDOpts{
		RegionID: idInt,
	}
	if categoryName != "" {
		opts.CategoryName = &categoryName
	}
	if groupCategoryName != "" {
		opts.GroupCategoryName = &groupCategoryName
	}

	categories, err := c.service.ListCategoriesByRegionID(ctx.Request().Context(), opts)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, categories)
}
