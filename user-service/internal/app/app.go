package app

import (
	"context"
	"fmt"
	"github.com/BeksultanSE/Assignment1-user/config"
	"github.com/BeksultanSE/Assignment1-user/internal/adapter/grpc"
	"github.com/BeksultanSE/Assignment1-user/internal/adapter/mongo"
	"github.com/BeksultanSE/Assignment1-user/internal/usecase"
	"github.com/BeksultanSE/Assignment1-user/pkg/hashing"
	mongoConn "github.com/BeksultanSE/Assignment1-user/pkg/mongo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const serviceName = "user-service"

type App struct {
	grpcServer *grpc.ServerAPI
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	log.Printf(fmt.Sprintf("Initializing %s service...", serviceName))

	log.Println("Connecting to DB:", cfg.Mongo.Database)
	mongoDB, err := mongoConn.NewDB(ctx, cfg.Mongo)
	if err != nil {
		return nil, fmt.Errorf("error connecting to DB: %v", err)
	}

	aiRepo := mongo.NewAutoInc(mongoDB.Conn)
	userRepo := mongo.NewUserRepo(mongoDB.Conn)

	hasher := hashing.NewBcryptHasher()

	userUsecase := usecase.NewUserUsecase(aiRepo, userRepo, hasher)

	grpcServer := grpc.New(cfg.Server, userUsecase)

	app := &App{
		grpcServer: grpcServer,
	}

	return app, nil
}

func (app *App) Start() error {
	errCh := make(chan error)

	app.grpcServer.Run(errCh)

	log.Printf(fmt.Sprintf("Starting %s service...", serviceName))

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case errRun := <-errCh:
		return errRun
	case sig := <-shutdownCh:
		log.Printf(fmt.Sprintf("Received shutdown signal: %s", sig.String()))
		app.Stop()
		log.Printf(fmt.Sprintf("Stopping %s service...", serviceName))
	}
	return nil
}

func (app *App) Stop() {
	err := app.grpcServer.Stop()
	if err != nil {
		log.Printf(fmt.Sprintf("Error stopping %s service: %v", serviceName, err))
	}
}
