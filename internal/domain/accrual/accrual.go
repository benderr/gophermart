package accrual

import "errors"

type Status string

const (
	REGISTERED Status = "REGISTERED"
	PROCESSING Status = "PROCESSING"
	INVALID    Status = "INVALID"
	PROCESSED  Status = "PROCESSED"
)

type Order struct {
	Order   string   `json:"order"`
	Status  Status   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

type RegisterOrder struct {
	Order string `json:"order"`
	Goods []Good `json:"goods"`
}

type Good struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

var (
	ErrCastError    = errors.New("accrual service cast error")
	ErrUnregistered = errors.New("unregistered order")
)
