package delivery

import (
	"context"
	"sync"
	"time"

	"github.com/benderr/gophermart/internal/domain/orders"
	"github.com/benderr/gophermart/internal/logger"
)

type AccrualUsecase interface {
	GetProcessOrders(ctx context.Context) ([]orders.Order, error)
	CheckOrder(ctx context.Context, order string) error
}

type processOrdersTask struct {
	accrual       AccrualUsecase
	workPoolLimit int
	logger        logger.Logger
}

func New(ac AccrualUsecase, logger logger.Logger, workPoolLimit int) *processOrdersTask {
	return &processOrdersTask{
		accrual:       ac,
		workPoolLimit: workPoolLimit,
		logger:        logger,
	}
}

func (p *processOrdersTask) runIteration(ctx context.Context) error {
	list, err := p.accrual.GetProcessOrders(ctx)
	p.logger.Infoln("[START ITERATION]")
	p.logger.Infoln("[FETCHED ORDERS]", list)

	if err != nil {
		p.logger.Errorln("fetched orders errors", err)
		return err
	}

	count := len(list)
	if count == 0 {
		return nil
	}

	jobs := make(chan *orders.Order, count)
	results := make(chan error, count)
	wg := &sync.WaitGroup{}

	wg.Add(p.workPoolLimit)

	for i := 0; i < p.workPoolLimit; i++ {
		p.logger.Infoln("[START WORKER]", i)
		i := i
		go worker(i, p.logger, wg, jobs, results, func(o *orders.Order) error {
			p.logger.Infow("[CHECK ORDER START]", "order", o.Number, "job", i)
			err := p.accrual.CheckOrder(ctx, o.Number)
			if err != nil {
				p.logger.Errorln("[CHECK ORDER FAILED]", i, o.Number, err)
			}
			return err
		})
	}

	for _, m := range list {
		p.logger.Infoln("[SENT JOB]", m)
		m := m
		jobs <- &m
	}

	close(jobs)

	wg.Wait()
	close(results)
	p.logger.Infoln("[FINISH ITERATION]")
	return nil
}

func (p *processOrdersTask) Run(ctx context.Context, interval int) {
	go func() {
		p.runIteration(ctx)
		ticker := time.NewTicker(time.Second * time.Duration(interval))
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.runIteration(ctx)
			}
		}
	}()
}

type WorkerFunc func(*orders.Order) error

func worker(id int, logger logger.Logger, wg *sync.WaitGroup, jobs <-chan *orders.Order, results chan<- error, fn WorkerFunc) {
	defer wg.Done()
	for m := range jobs {
		logger.Infoln("[GET FROM JOB]", m)
		err := fn(m)
		results <- err
	}
}
