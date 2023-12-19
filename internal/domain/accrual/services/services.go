package services

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/benderr/gophermart/internal/domain/accrual"
	"github.com/benderr/gophermart/internal/logger"
	"github.com/go-resty/resty/v2"
)

type accrualService struct {
	client *resty.Client
	logger logger.Logger
}

func New(server string, logger logger.Logger) *accrualService {
	client := resty.
		New().
		SetBaseURL(server)

	setCustomRetries(client, 3)

	return &accrualService{
		client: client,
		logger: logger,
	}
}

func (a *accrualService) GetOrder(number string) (*accrual.Order, error) {
	resp, err := a.client.R().SetHeader("Content-Type", "application/json").
		SetResult(&accrual.Order{}).
		Get(fmt.Sprintf("/api/orders/%s", number))

	if err != nil {
		a.logger.Errorln("[GET ORDER FAILED]", err)
		return nil, err
	}

	if resp.StatusCode() == http.StatusNoContent {
		a.logger.Infoln("[ORDER NO CONTENT]", number)
		return nil, accrual.ErrUnregistered
	}

	if order, ok := resp.Result().(*accrual.Order); ok {
		return order, nil
	}

	return nil, accrual.ErrCastError
}

// Для проверки корректности работы делаем метод регистрации заказа в сервисе accrual
func (a *accrualService) Registration(number string) error {
	goods := make([]accrual.Good, 0)
	goods = append(goods, accrual.Good{Price: rand.Float64() * 100, Description: "Стиральная машинка LG"})
	goods = append(goods, accrual.Good{Price: rand.Float64() * 100, Description: "second good desc"})

	order := &accrual.RegisterOrder{
		Order: number,
		Goods: goods,
	}

	r, err := a.client.R().SetHeader("Content-Type", "application/json").
		SetBody(order).
		Post("/api/orders")

	if err != nil {
		a.logger.Infoln("[REG ORDER FAILED]", err)
	}

	if r != nil {
		a.logger.Infoln("[REG ORDER RESULT]", number, r.StatusCode(), string(r.Body()))
	}

	return err
}

func setCustomRetries(client *resty.Client, count int) {
	client.SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second).
		SetRetryCount(count).
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			retryAfter := resp.Header().Get("Retry-After")
			if len(retryAfter) == 0 {
				wait, err := strconv.Atoi(retryAfter)
				if err != nil {
					return 0, errors.New("quota exceeded")
				}
				if wait > 0 {
					return time.Duration(wait) * time.Second, nil
				}
			}

			return 0, errors.New("quota exceeded")
		})
}
