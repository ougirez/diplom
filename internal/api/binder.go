package api

import (
	"bytes"
	"fmt"
	"github.com/ougirez/diplom/internal/pkg/constants"

	"github.com/labstack/gommon/log"

	"github.com/bytedance/sonic"
	"github.com/labstack/echo/v4"
)

type binderImpl struct{}

func (b *binderImpl) Bind(i interface{}, ctx echo.Context) error {
	db := &echo.DefaultBinder{}

	buf := &bytes.Buffer{}
	_, err := buf.ReadFrom(ctx.Request().Body)
	if err != nil {
		log.Errorf("Unmarshal error: %s", err)
		return err
	}

	err = sonic.Unmarshal(buf.Bytes(), i)
	if err != nil {
		log.Errorf("Unmarshal error: %s", err)
	}

	if err := db.BindQueryParams(ctx, i); err != nil {
		return fmt.Errorf("%w: %v", constants.ErrBindRequest, err)
	}

	if err := db.BindPathParams(ctx, i); err != nil {
		return fmt.Errorf("%w: %v", constants.ErrBindRequest, err)
	}

	if err := db.BindHeaders(ctx, i); err != nil {
		return fmt.Errorf("%w: %v", constants.ErrBindRequest, err)
	}

	if err := ctx.Validate(i); err != nil {
		return fmt.Errorf("%w: %v", constants.ErrValidateRequest, err)
	}

	return nil
}

func NewBinder() echo.Binder {
	return &binderImpl{}
}
