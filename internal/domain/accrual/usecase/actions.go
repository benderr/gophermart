package usecase

import (
	"context"
	"errors"

	"github.com/benderr/gophermart/internal/domain/accrual"
	"github.com/benderr/gophermart/internal/domain/orders"
)

type accrualUsecase struct {
	orderRepo      OrderRepo
	accrualService AccrualService
	orderUsecase   OrderUsecase
}

func New(op OrderRepo, as AccrualService, ou OrderUsecase) *accrualUsecase {
	return &accrualUsecase{
		orderRepo:      op,
		accrualService: as,
		orderUsecase:   ou,
	}
}

func (a *accrualUsecase) GetProcessOrders(ctx context.Context) ([]orders.Order, error) {
	list, err := a.orderRepo.GetOrdersByStatuses(ctx, orders.NEW, orders.PROCESSING)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (a *accrualUsecase) CheckOrder(ctx context.Context, order string) error {
	info, err := a.accrualService.GetOrder(order)

	//Этот кусок для удобства тестирования
	if err != nil && errors.Is(err, accrual.ErrUnregistered) {
		a.accrualService.Registration(order)
	}

	if err != nil {
		return err
	}

	return a.orderUsecase.ChangeStatus(ctx, info.Order, orders.Status(info.Status), info.Accrual)
}
