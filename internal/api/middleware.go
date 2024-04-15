package api

import (
	"github.com/labstack/echo/v4"
	"github.com/ougirez/diplom/internal/pkg/constants"
	"github.com/ougirez/diplom/internal/pkg/utils"
	"github.com/spf13/viper"
)

func (svc *APIService) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		//cookie, err := ctx.Cookie(constants.CookieKeyAuthToken)
		//if err != nil {
		//	return constants.ErrMissingAuthCookie
		//}

		//token, err := utils.ParseAuthToken(cookie.Value)
		//if err != nil {
		//	return err
		//}

		//if status, err := svc.store.GetUserStatus(ctx.Request().Context(), token.UserID); err != nil {
		//	return err
		//} else if status != string(core.UserStatusApproved) {
		//	return constants.ErrUnauthorized
		//}

		//ctx.Set(constants.CtxKeyUserID, token.UserID)

		return next(ctx)
	}
}

func (svc *APIService) AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		cookie, err := ctx.Cookie(constants.CookieKeySecretToken)
		if err != nil {
			return constants.ErrUnauthorized
		}

		token, err := utils.ParseAuthToken(cookie.Value)
		if err != nil {
			return err
		}

		if token.Secret != viper.GetString(constants.ViperSecretKey) {
			return constants.ErrUnauthorized
		}

		return next(ctx)
	}
}
