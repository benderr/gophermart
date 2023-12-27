package storage

import (
	"context"
	"database/sql"
	"os"

	"github.com/benderr/gophermart/internal/config"
	"github.com/benderr/gophermart/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func MustLoad(ctx context.Context, config *config.Config, logger logger.Logger) *sql.DB {
	if len(config.DatabaseDsn) == 0 {
		logger.Errorln("[DB]: database dsn not specified")
		panic("database dsn not specified")
	}

	db, dberr := sql.Open("pgx", config.DatabaseDsn)
	if dberr != nil {
		logger.Errorln("[DB]: database dsn not specified", dberr)
		db.Close()
		panic(dberr)
	}

	err := runMigration(ctx, db)

	if err != nil {
		logger.Errorln("[DB]: migration failed", dberr)
		db.Close()
		panic(err)
	}

	return db
}

func runMigration(ctx context.Context, db *sql.DB) error {

	content, err := os.ReadFile("./init.sql")

	if err != nil {
		return err
	}

	sql := string(content)

	_, err = db.ExecContext(ctx, sql)

	if err != nil {
		return err
	}

	return nil
}
