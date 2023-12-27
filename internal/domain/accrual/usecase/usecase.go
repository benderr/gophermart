package usecase

import (
	"context"

	"github.com/benderr/gophermart/internal/domain/accrual"
	"github.com/benderr/gophermart/internal/domain/orders"
)

type OrderRepo interface {
	GetOrdersByStatuses(ctx context.Context, status ...orders.Status) ([]orders.Order, error)
}

type AccrualService interface {
	GetOrder(number string) (*accrual.Order, error)
	Registration(number string) error
}

type OrderUsecase interface {
	ChangeStatus(ctx context.Context, number string, status orders.Status, accrual *float64) error
}
