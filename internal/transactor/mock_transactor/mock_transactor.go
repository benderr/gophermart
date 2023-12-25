package mocktransactor

import (
	"context"
	"database/sql"
)

type MockTransactor struct {
}

func New() *MockTransactor {
	return &MockTransactor{}
}

func (m *MockTransactor) Within(ctx context.Context, tFunc func(ctx context.Context, tx *sql.Tx) error) error {
	return tFunc(ctx, nil)
}
