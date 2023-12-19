package main

import (
	"context"
	"net/http"

	"github.com/benderr/gophermart/internal/app"
	"github.com/benderr/gophermart/internal/config"

	"github.com/benderr/gophermart/internal/logger"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func main() {
	conf := config.MustLoad()
	ctx := context.Background()

	logger, sync := logger.New()
	defer sync()

	e := app.Run(ctx, conf, logger)

	e.Logger.Fatal(e.Start(string(conf.Server)))
}
