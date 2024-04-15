package api

import (
	"errors"
	"github.com/ougirez/diplom/internal/domain"
	"github.com/ougirez/diplom/internal/pkg/constants"
	"net/http"

	"github.com/labstack/echo/v4"
)

func httpErrorHandler(err error, c echo.Context) {
	msg := err.Error()
	code := http.StatusInternalServerError
	for err != nil {
		if ce, ok := err.(*constants.CodedError); ok {
			code = ce.Code()
			break
		}
		err = errors.Unwrap(err)
	}

	_ = c.JSON(code, domain.ErrorResponse{
		Message: msg,
		Code:    code,
	})
}
