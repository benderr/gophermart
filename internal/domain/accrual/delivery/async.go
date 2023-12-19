package delivery

import (
	"context"

	"github.com/benderr/gophermart/internal/domain/orders"
	"github.com/benderr/gophermart/internal/logger"
)

type AccrualUsecase interface {
	GetProcessOrders(ctx context.Context) ([]orders.Order, error)
}

type Publisher interface {
	Publish(topic string, payload any) error
}
type processOrdersTask struct {
	accrual   AccrualUsecase
	logger    logger.Logger
	publisher Publisher
}

func New(ac AccrualUsecase, publisher Publisher, logger logger.Logger) *processOrdersTask {
	return &processOrdersTask{
		accrual:   ac,
		logger:    logger,
		publisher: publisher,
	}
}

func (p *processOrdersTask) Run(ctx context.Context) error {
	list, err := p.accrual.GetProcessOrders(ctx)

	p.logger.Infoln("[START PUBLISH]")
	p.logger.Infoln("[FETCHED ORDERS]", list)

	if err != nil {
		p.logger.Errorln("fetched orders errors", err)
		return err
	}

	count := len(list)
	if count == 0 {
		return nil
	}

	for _, m := range list {
		p.logger.Infoln("[SENT JOB]", m)
		m := m
		p.publisher.Publish("order.check", &m)
	}

	p.logger.Infoln("[FINISH PUBLISH]")

	return nil
}
