package api

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ougirez/diplom/internal/api/controller"
	"github.com/ougirez/diplom/internal/pkg/logger"
	"github.com/ougirez/diplom/internal/pkg/store"
	"github.com/ougirez/diplom/internal/service"
)

type APIService struct {
	router           *echo.Echo
	providersService *service.Service
}

func (svc *APIService) Serve(addr string) {
	logger.Fatal(context.Background(), svc.router.Start(addr))
}

func (svc *APIService) Shutdown(ctx context.Context) error {
	return svc.router.Shutdown(ctx)
}

func NewAPIService(store store.Store) (*APIService, error) {
	svc := &APIService{router: echo.New()}

	svc.router.Validator = NewValidator()
	svc.router.Binder = NewBinder()
	svc.router.Use(middleware.Logger())
	svc.router.HTTPErrorHandler = httpErrorHandler
	svc.router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},                    // Разрешить запросы только от этого домена
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE}, // Разрешить эти HTTP-методы
		AllowHeaders: []string{"Content-Type", "Authorization"},            // Разрешить эти заголовки
	}))

	svc.providersService = service.NewProvidersService(store)

	api := svc.router.Group("/api/v1")
	cntrl := controller.NewController(svc.providersService)

	fgbu := api.Group("/providers")
	fgbu.POST("/fgbu/backfill", cntrl.BackFillFGBUData)

	regions := api.Group("/regions")
	regions.GET("/list", cntrl.GetRegions)
	regions.GET("/category", cntrl.GetCategoryDataByRegions)
	regions.GET("/:id/categories", cntrl.GetCategoriesByRegionID)

	return svc, nil
}
