package api

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ougirez/diplom/internal/api/controller"
	"github.com/ougirez/diplom/internal/pkg/logger"
	"github.com/ougirez/diplom/internal/pkg/store"
	"github.com/ougirez/diplom/internal/service/providers"
)

type APIService struct {
	router           *echo.Echo
	providersService *providers.Service
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

	svc.providersService = providers.NewProvidersService(store)

	api := svc.router.Group("/api/v1")
	cntrl := controller.NewController(svc.providersService)

	fgbu := api.Group("/providers")
	fgbu.POST("/fgbu/backfill", cntrl.BackFillFGBUData)

	regions := api.Group("/regions")
	regions.GET("/list", cntrl.GetProvidersRegions)
	regions.GET("/:id/categories", cntrl.GetCategoriesByRegionID)

	//auth := api.Group("/auth")
	//auth.POST("/signup", controller.SignupUser)
	//auth.POST("/login", controller.LoginUser)
	//auth.DELETE("/logout", controller.LogoutUser)
	//
	//user := api.Group("/user", svc.AuthMiddleware)
	//user.GET("/get", controller.GetUser)
	//
	//account := api.Group("/accounts", svc.AuthMiddleware)
	//
	//account.POST("/create", controller.CreateAccount)
	//account.GET("/list", controller.ListUserAccounts)
	//account.POST("/refill", controller.RefillAccount)
	//account.POST("/withdraw", controller.WithdrawFromAccount)
	//account.POST("/buy", controller.MakePurchase)
	//account.POST("/sell", controller.MakeSale)
	//
	//oauth := api.Group("/oauth")
	//oauth.GET("/telegram/:user_id", controller.OAuthTelegram, svc.OAuthTelegramMiddleware)
	//
	//admin := api.Group("/admin")
	//admin.POST("/login", controller.LoginAdmin)
	//admin.POST("/update_user_status", controller.UpdateUserStatus, svc.AdminMiddleware)
	//admin.POST("/list_users", controller.ListUsers, svc.AdminMiddleware)
	//
	//currencies := api.Group("/currencies", svc.AuthMiddleware)
	//currencies.GET("/list", controller.ListCurrencies)
	//currencies.GET("/data", controller.GetCurrencyData)
	//
	//operations := api.Group("/operations", svc.AuthMiddleware)
	//operations.POST("/list", controller.ListOperations)
	//
	//funtik := api.Group("/funtik", svc.AuthMiddleware)
	//funtik.POST("/subscribe", controller.SubscribeToFuntik)

	return svc, nil
}
