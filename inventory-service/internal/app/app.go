package app

import (
	"context"
	"fmt"
	"github.com/BeksultanSE/Assignment1-inventory/config"
	grpcAPI "github.com/BeksultanSE/Assignment1-inventory/internal/adapter/grpc"
	//httpRepo "github.com/BeksultanSE/Assignment1-inventory/internal/adapter/http"
	mongoRepo "github.com/BeksultanSE/Assignment1-inventory/internal/adapter/mongo"
	"github.com/BeksultanSE/Assignment1-inventory/internal/usecase"
	mongoConn "github.com/BeksultanSE/Assignment1-inventory/pkg/mongo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const serviceName = "inventory-service"

type App struct {
	//httpServer *httpRepo.API
	grpcServer *grpcAPI.ServerAPI
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	log.Printf(fmt.Sprintf("Initializing %s service...", serviceName))

	log.Println("Connecting to DB:", cfg.Mongo.Database)
	mongoDB, err := mongoConn.NewDB(ctx, cfg.Mongo)
	if err != nil {
		return nil, fmt.Errorf("error connecting to DB: %v", err)
	}

	aiRepo := mongoRepo.NewAutoInc(mongoDB.Conn)
	pRepo := mongoRepo.NewProductRepo(mongoDB.Conn)

	pUsecase := usecase.NewProduct(aiRepo, pRepo)

	//httpServer := httpRepo.New(cfg.Server, pUsecase)
	grpcServer := grpcAPI.New(cfg.Server, pUsecase)

	app := &App{
		//httpServer: httpServer,
		grpcServer: grpcServer,
	}

	return app, nil
}

func (app *App) Start() error {
	errCh := make(chan error)

	//app.httpServer.Run(errCh)
	app.grpcServer.Run(errCh)

	log.Printf(fmt.Sprintf("Starting %s service...", serviceName))

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case errRun := <-errCh:
		return errRun
	case sig := <-shutdownCh:

		log.Printf(fmt.Sprintf("Received %v signal, shutting down...", sig))
		app.Stop()
		log.Println("graceful shutdown completed!")
	}
	return nil
}

func (app *App) Stop() {
	//err := app.httpServer.Stop()
	err := app.grpcServer.Stop()
	if err != nil {
		log.Println("failed to shutdown http service:", err)
	}
}
