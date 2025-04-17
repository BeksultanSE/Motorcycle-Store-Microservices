package main

import (
	"context"
	"github.com/BeksultanSE/Assignment1-api-gateway/config"
	"github.com/BeksultanSE/Assignment1-api-gateway/internal/app"
	"log"
)

func main() {
	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	if err := application.Start(); err != nil {
		log.Fatalf("Failed to start app: %v", err)
	}
}
