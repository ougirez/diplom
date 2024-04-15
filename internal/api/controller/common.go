package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/ougirez/diplom/internal/pkg/constants"
)

func GetUserIDFromCtx(ctx echo.Context) (int64, error) {
	id, ok := ctx.Get(constants.CtxKeyUserID).(int64)
	if !ok {
		return 0, constants.ErrUnauthorized
	}
	return id, nil
}
