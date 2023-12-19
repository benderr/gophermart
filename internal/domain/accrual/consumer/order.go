package consumer

import (
	"context"
	"errors"

	"github.com/benderr/gophermart/internal/domain/orders"
)

type Consumer interface {
	Consume(topic string, callback func(ctx context.Context, payload any) error)
}

type AccrualUsecase interface {
	CheckOrder(ctx context.Context, order string) error
}

func RegisterHandler(au AccrualUsecase, consumer Consumer) {
	consumer.Consume("order.check", func(ctx context.Context, payload any) error {
		if v, ok := payload.(*orders.Order); ok {
			return au.CheckOrder(ctx, v.Number)
		}
		return errors.New("consumer cast error")
	})
}
