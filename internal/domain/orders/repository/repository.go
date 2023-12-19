package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/benderr/gophermart/internal/domain/orders"
	"github.com/benderr/gophermart/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type orderRepository struct {
	db  *sql.DB
	log logger.Logger
}

func New(db *sql.DB, log logger.Logger) *orderRepository {
	return &orderRepository{db: db, log: log}
}

func (u *orderRepository) GetByNumber(ctx context.Context, number string) (*orders.Order, error) {

	row := u.db.QueryRowContext(ctx, "SELECT order_num, user_id, status, accrual, uploaded_at from orders WHERE order_num = $1", number)
	var ord orders.Order
	err := row.Scan(&ord.Number, &ord.UserID, &ord.Status, &ord.Accrual, &ord.UploadedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, orders.ErrNotFound
		}

		return nil, err
	}

	return &ord, nil
}

func (u *orderRepository) GetOrdersByStatuses(ctx context.Context, statuses ...orders.Status) ([]orders.Order, error) {
	orderlist := make([]orders.Order, 0)

	statusarr, params := createInParams[orders.Status](statuses)
	u.log.Infoln("configured sql", statusarr, params)
	rows, err := u.db.QueryContext(ctx, `SELECT order_num, status, accrual, user_id, uploaded_at from orders WHERE status in (`+params+`)  ORDER BY uploaded_at desc`, statusarr...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var v orders.Order
		err = rows.Scan(&v.Number, &v.Status, &v.Accrual, &v.UserID, &v.UploadedAt)
		if err != nil {
			return nil, err
		}

		orderlist = append(orderlist, v)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return orderlist, nil
}

func (u *orderRepository) GetOrdersByUser(ctx context.Context, userid string) ([]orders.Order, error) {
	orderlist := make([]orders.Order, 0)

	rows, err := u.db.QueryContext(ctx, "SELECT order_num, status, accrual, user_id, uploaded_at from orders WHERE user_id=$1 ORDER BY uploaded_at desc", userid)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var v orders.Order
		err = rows.Scan(&v.Number, &v.Status, &v.Accrual, &v.UserID, &v.UploadedAt)
		if err != nil {
			return nil, err
		}

		orderlist = append(orderlist, v)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return orderlist, nil
}

func (u *orderRepository) Create(ctx context.Context, userid string, number string, status orders.Status) (*orders.Order, error) {
	_, err := u.db.ExecContext(ctx, `INSERT INTO orders (user_id, order_num, status) VALUES ($1, $2, $3)`, userid, number, status)
	if err != nil {
		var perr *pgconn.PgError
		if errors.As(err, &perr) && perr.Code == pgerrcode.UniqueViolation {
			return nil, orders.ErrAlreadyExist
		}

		return nil, err
	}

	created, err := u.GetByNumber(ctx, number)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (u *orderRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, number string, status orders.Status) error {
	_, err := tx.ExecContext(ctx, `UPDATE orders SET status=$1 WHERE order_num=$2`, status, number)
	return err
}

func (u *orderRepository) UpdateAccrual(ctx context.Context, tx *sql.Tx, number string, accrual *float64) error {
	_, err := tx.ExecContext(ctx, `UPDATE orders SET accrual=$1 WHERE order_num=$2`, accrual, number)
	return err
}

// создаем параметры для выражения IN
// Пример для []int{1,2,3}
// Результат []any{1,2,3} "$1,$2,$3"
func createInParams[T comparable](arr []T) ([]any, string) {
	res := make([]any, 0)
	params := make([]string, 0)
	for i, v := range arr {
		res = append(res, v)
		params = append(params, fmt.Sprintf("$%v", i+1))

	}
	return res, strings.Join(params, ",")
}
