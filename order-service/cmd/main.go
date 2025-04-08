package main

import (
	"context"
	"github.com/BeksultanSE/Assignment1-order/config"
	"github.com/BeksultanSE/Assignment1-order/internal/app"
	"log"
)

func main() {
	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		log.Printf("error loading config: %v", err)
		return
	}

	app, err := app.New(ctx, cfg)
	if err != nil {
		log.Printf("error creating app: %v", err)
		return
	}

	err = app.Start()
	if err != nil {
		log.Printf("error starting app: %v", err)
		return
	}
}
