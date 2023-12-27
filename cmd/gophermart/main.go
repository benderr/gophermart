package main

import (
	"context"

	"github.com/benderr/gophermart/internal/app"
	"github.com/benderr/gophermart/internal/config"
)

func main() {
	conf := config.MustLoad()
	ctx := context.Background()

	app.Run(ctx, conf)
}
