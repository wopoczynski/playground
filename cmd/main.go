package main

import (
	"context"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"

	server "github.com/wopoczynski/playground/internal/application"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = godotenv.Load()

	var cfg server.Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		panic("app config boot failed")
	}

	app, err := server.New(cfg)
	if err != nil {
		panic(fmt.Errorf("server error %w", err))
	}

	app.Init(ctx)
	app.Start(ctx)
}
