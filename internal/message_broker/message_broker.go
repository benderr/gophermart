package messagebroker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/benderr/gophermart/internal/logger"
)

type message struct {
	topic   string
	payload any
}

type callback func(ctx context.Context, payload any) error

type messageBroker struct {
	workPoolLimit int
	consumers     map[string][]callback
	messages      chan *message
	logger        logger.Logger
	mu            sync.Mutex
}

func New(workPoolLimit int, logger logger.Logger) *messageBroker {
	return &messageBroker{
		consumers:     make(map[string][]callback),
		messages:      make(chan *message, workPoolLimit),
		workPoolLimit: workPoolLimit,
		logger:        logger,
	}
}

func (m *messageBroker) Run(ctx context.Context) {
	for i := 0; i < m.workPoolLimit; i++ {
		m.logger.Infoln("[RUN BROKER WORKER]", i)
		go m.listenMessages(ctx, i)
	}
}

func (m *messageBroker) listenMessages(ctx context.Context, i int) {
	for {
		select {
		case <-ctx.Done():
			return
		case v, opened := <-m.messages:
			if !opened {
				return
			}
			if consumers, ok := m.consumers[v.topic]; ok {
				for _, consumer := range consumers {
					m.logger.Infoln("[BROKER CONSUME]", i)
					err := retrableDo(func() error {
						return consumer(ctx, v.payload)
					})
					if err != nil {
						m.logger.Infoln("[BROKER ERROR]", err)
					}
				}
			}
		}
	}
}

func (m *messageBroker) Publish(topic string, payload any) error {
	m.logger.Infow("[BROKER MESSAGE]", "topic", topic, "payload", payload)
	m.messages <- &message{topic: topic, payload: payload}
	return nil
}

func (m *messageBroker) Consume(topic string, cb func(ctx context.Context, payload any) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Infoln("[BROKER REGISTER CONSUME]", topic)
	if _, ok := m.consumers[topic]; !ok {
		m.consumers[topic] = make([]callback, 0)
	}
	m.consumers[topic] = append(m.consumers[topic], cb)
}

func retrableDo(fn func() error) error {
	attempt := 0
	allErrors := make([]error, 0)
	for {
		err := fn()
		if err == nil {
			return nil
		}
		allErrors = append(allErrors, err)
		attempt++

		if !retryCondition(attempt, err) {
			return errors.Join(allErrors...)
		}
	}
}

func retryCondition(attempt int, err error) bool {
	if attempt < 7 {
		time.Sleep(time.Millisecond * 100 * time.Duration(attempt))
		return true
	}
	return false
}
