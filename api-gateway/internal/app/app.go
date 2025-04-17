package app

import (
	"context"
	"fmt"
	"github.com/BeksultanSE/Assignment1-api-gateway/config"
	"github.com/BeksultanSE/Assignment1-api-gateway/internal/adapter/grpc"
	"github.com/BeksultanSE/Assignment1-api-gateway/internal/adapter/http"
	handlers "github.com/BeksultanSE/Assignment1-api-gateway/internal/adapter/http/handler"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const ServiceName = "api-gateway"

type App struct {
	cfg         *config.Config
	grpcClients *grpc.Clients
	httpServer  *http.Server
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	log.Printf("Initializing %s service...", ServiceName)

	grpcClients, err := grpc.NewClients(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC clients: %w", err)
	}

	handler := handlers.NewHandler(grpcClients)
	httpServer := http.NewServer(*cfg, handler)

	app := &App{
		cfg:         cfg,
		grpcClients: grpcClients,
		httpServer:  httpServer,
	}

	return app, nil
}

func (app *App) Start() error {
	errCh := make(chan error)

	app.httpServer.Run(errCh)

	log.Printf("Starting %s service...", ServiceName)

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case errRun := <-errCh:
		return errRun
	case sig := <-shutdownCh:
		log.Printf("Received %v signal, shutting down...", sig)
		app.Stop()
		log.Println("Graceful shutdown completed!")
	}
	return nil
}

func (app *App) Stop() {
	err := app.httpServer.Stop()
	if err != nil {
		log.Printf("Failed to stop http server: %v", err)
	}
	app.grpcClients.Close()
}
