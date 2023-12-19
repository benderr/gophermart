package main

import (
	"context"
	"net/http"

	"github.com/benderr/gophermart/internal/config"

	accrualdelivery "github.com/benderr/gophermart/internal/domain/accrual/delivery"
	"github.com/benderr/gophermart/internal/domain/accrual/services"
	accrualusecase "github.com/benderr/gophermart/internal/domain/accrual/usecase"
	userdelivery "github.com/benderr/gophermart/internal/domain/user/delivery"
	userrepo "github.com/benderr/gophermart/internal/domain/user/repository"
	userusecase "github.com/benderr/gophermart/internal/domain/user/usecase"
	"github.com/benderr/gophermart/internal/transactor"

	orderdelivery "github.com/benderr/gophermart/internal/domain/orders/delivery"
	orderrepo "github.com/benderr/gophermart/internal/domain/orders/repository"
	orderusecase "github.com/benderr/gophermart/internal/domain/orders/usecase"

	balancedelivery "github.com/benderr/gophermart/internal/domain/balance/delivery"
	balancerepo "github.com/benderr/gophermart/internal/domain/balance/repository"
	balanceusecase "github.com/benderr/gophermart/internal/domain/balance/usecase"

	withdrawdelivery "github.com/benderr/gophermart/internal/domain/withdrawal/delivery"
	withdrawrepo "github.com/benderr/gophermart/internal/domain/withdrawal/repository"
	withdrawusecase "github.com/benderr/gophermart/internal/domain/withdrawal/usecase"

	"github.com/benderr/gophermart/internal/logger"
	"github.com/benderr/gophermart/internal/session"
	"github.com/benderr/gophermart/internal/storage"
	"github.com/benderr/gophermart/internal/utils"
	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	p, _ := utils.HashPassword("123")

	logger.Infoln("pass", p)

	db := storage.MustLoad(ctx, conf, logger)

	sessionManager := session.New(conf.SecretKey)
	trsctr := transactor.New(db)

	userRepo := userrepo.New(db, logger)
	orderRepo := orderrepo.New(db, logger)
	balanceRepo := balancerepo.New(db, logger)
	withdrawRepo := withdrawrepo.New(db, logger)

	accrualService := services.New(string(conf.AccrualServer), logger)

	userUsecase := userusecase.New(userRepo, logger)
	orderUsecase := orderusecase.New(orderRepo, balanceRepo, trsctr, logger)
	balanceUsecase := balanceusecase.New(balanceRepo, withdrawRepo, trsctr, logger)
	withdrawUsecase := withdrawusecase.New(withdrawRepo, logger)
	accrualUsecase := accrualusecase.New(orderRepo, accrualService, orderUsecase)

	acrualTask := accrualdelivery.New(accrualUsecase, logger, 5)

	e := echo.New()
	validate := validator.New()
	// validate.RegisterValidation("moon", func(fl validator.FieldLevel) bool {
	// 	err := utils.ValidateMoon(fl.Field().String())
	// 	if err != nil {
	// 		return false
	// 	}
	// 	return true
	// })
	e.Validator = &CustomValidator{validator: validate}
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	publicGroup := e.Group("")

	privateGroup := e.Group("", echojwt.WithConfig(echojwt.Config{
		SigningKey:    []byte(conf.SecretKey),
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(session.UserClaims) },
		ErrorHandler: func(c echo.Context, err error) error {
			logger.Errorln(err)
			return err
		},
	}))

	userdelivery.NewUserHandlers(publicGroup, userUsecase, sessionManager, logger)
	orderdelivery.NewOrderHandlers(privateGroup, orderUsecase, sessionManager, logger)
	balancedelivery.NewBalanceHandlers(privateGroup, balanceUsecase, sessionManager, logger)
	withdrawdelivery.NewWithdrawHandlers(privateGroup, withdrawUsecase, sessionManager, logger)

	acrualTask.Run(ctx, 5)

	e.Logger.Fatal(e.Start(string(conf.Server)))
}
