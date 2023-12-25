package app

import (
	"context"
	"net/http"

	"github.com/benderr/gophermart/internal/config"
	messageBroker "github.com/benderr/gophermart/internal/messagebroker"

	accrualConsumer "github.com/benderr/gophermart/internal/domain/accrual/consumer"
	accrualDelivery "github.com/benderr/gophermart/internal/domain/accrual/delivery"
	acrualService "github.com/benderr/gophermart/internal/domain/accrual/services"
	accrualUsecase "github.com/benderr/gophermart/internal/domain/accrual/usecase"
	userDelivery "github.com/benderr/gophermart/internal/domain/user/delivery"
	userRepository "github.com/benderr/gophermart/internal/domain/user/repository"
	userUsecase "github.com/benderr/gophermart/internal/domain/user/usecase"
	"github.com/benderr/gophermart/internal/transactor"

	orderDelivery "github.com/benderr/gophermart/internal/domain/orders/delivery"
	orderRepository "github.com/benderr/gophermart/internal/domain/orders/repository"
	orderUsecase "github.com/benderr/gophermart/internal/domain/orders/usecase"

	balanceDelivery "github.com/benderr/gophermart/internal/domain/balance/delivery"
	balanceRepository "github.com/benderr/gophermart/internal/domain/balance/repository"
	balanceUsecase "github.com/benderr/gophermart/internal/domain/balance/usecase"

	withdrawDelivery "github.com/benderr/gophermart/internal/domain/withdrawal/delivery"
	withdrawRepository "github.com/benderr/gophermart/internal/domain/withdrawal/repository"
	withdrawUsecase "github.com/benderr/gophermart/internal/domain/withdrawal/usecase"

	"github.com/benderr/gophermart/internal/logger"
	"github.com/benderr/gophermart/internal/session"
	"github.com/benderr/gophermart/internal/storage"
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

func Run(ctx context.Context, conf *config.Config) {
	logger, sync := logger.New()
	defer sync()

	db := storage.MustLoad(ctx, conf, logger)

	sessionManager := session.New(conf.SecretKey)
	trsctr := transactor.New(db)
	msgBroker := messageBroker.New(5, logger)
	msgBroker.Run(ctx)

	userRepo := userRepository.New(db, logger)
	orderRepo := orderRepository.New(db, logger)
	balanceRepo := balanceRepository.New(db, logger)
	withdrawRepo := withdrawRepository.New(db, logger)
	accrualSrv := acrualService.New(string(conf.AccrualServer), logger)

	userUsecase := userUsecase.New(userRepo, logger)
	orderUsecase := orderUsecase.New(orderRepo, balanceRepo, trsctr, msgBroker, logger)
	balanceUsecase := balanceUsecase.New(balanceRepo, withdrawRepo, trsctr, logger)
	withdrawUsecase := withdrawUsecase.New(withdrawRepo, logger)
	accrualUsecase := accrualUsecase.New(orderRepo, accrualSrv, orderUsecase, logger)

	accrualConsumer.RegisterHandler(accrualUsecase, msgBroker)

	e := echo.New()
	validate := validator.New()

	e.Validator = &CustomValidator{validator: validate}
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	publicGroup := e.Group("")

	privateGroup := e.Group("", echojwt.WithConfig(echojwt.Config{
		SigningKey:    []byte(conf.SecretKey),
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(session.UserClaims) },
	}))

	userDelivery.NewUserHandlers(publicGroup, userUsecase, sessionManager, logger)
	orderDelivery.NewOrderHandlers(privateGroup, orderUsecase, sessionManager, logger)
	balanceDelivery.NewBalanceHandlers(privateGroup, balanceUsecase, sessionManager, logger)
	withdrawDelivery.NewWithdrawHandlers(privateGroup, withdrawUsecase, sessionManager, logger)

	acrualTask := accrualDelivery.New(accrualUsecase, msgBroker, logger)
	acrualTask.Run(ctx)

	e.Logger.Fatal(e.Start(string(conf.Server)))
}
