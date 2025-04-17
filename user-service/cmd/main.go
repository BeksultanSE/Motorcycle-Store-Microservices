package main

import (
	"context"
	"github.com/BeksultanSE/Assignment1-user/config"
	"github.com/BeksultanSE/Assignment1-user/internal/app"
	"log"
)

func main() {
	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		log.Println("error loading config", err)
		return
	}

	app, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Println("error initializing app", err)
		return
	}

	err = app.Start()
	
	if err != nil {
		log.Println("error starting app", err)
		return
	}

}
