package transactor

import (
	"context"
	"database/sql"
)

type transactorInstance struct {
	db *sql.DB
}

func New(db *sql.DB) *transactorInstance {
	return &transactorInstance{db}
}

func (t *transactorInstance) Within(ctx context.Context, tFunc func(ctx context.Context, tx *sql.Tx) error) error {
	tx, err := t.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	err = tFunc(ctx, tx)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
