package main

import (
	"context"

	"github.com/benderr/gophermart/internal/app"
	"github.com/benderr/gophermart/internal/config"

	"github.com/benderr/gophermart/internal/logger"
)

func main() {
	conf := config.MustLoad()
	ctx := context.Background()

	logger, sync := logger.New()
	defer sync()

	e := app.Run(ctx, conf, logger)

	e.Logger.Fatal(e.Start(string(conf.Server)))
}
