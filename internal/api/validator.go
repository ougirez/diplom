package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type validatorImpl struct {
	validator *validator.Validate
}

func (v *validatorImpl) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func NewValidator() echo.Validator {
	return &validatorImpl{validator: validator.New()}
}
