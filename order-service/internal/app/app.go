package app

import (
	"context"
	"fmt"
	"github.com/BeksultanSE/Assignment1-order/config"
	grpcAPI "github.com/BeksultanSE/Assignment1-order/internal/adapter/grpc"
	gclients "github.com/BeksultanSE/Assignment1-order/internal/adapter/grpc/clients"
	"github.com/BeksultanSE/Assignment1-order/internal/adapter/kafka"
	mongoRepo "github.com/BeksultanSE/Assignment1-order/internal/adapter/mongo"
	"github.com/BeksultanSE/Assignment1-order/internal/usecase"
	mongoConn "github.com/BeksultanSE/Assignment1-order/pkg/mongo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const serviceName = "order-service"

type App struct {
	//httpServer *httpRepo.API
	grpcServer  *grpcAPI.ServerAPI
	grpcClients *gclients.Clients
	kafkaProd   *kafka.Producer
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	log.Printf(fmt.Sprintf("Initializing %s service...", serviceName))

	log.Println("Connecting to DB:", cfg.Mongo.Database)
	mongoDB, err := mongoConn.NewDB(ctx, cfg.Mongo)
	if err != nil {
		return nil, fmt.Errorf("error connecting to DB: %v", err)
	}

	aiRepo := mongoRepo.NewAutoInc(mongoDB.Conn)
	orderRepo := mongoRepo.NewOrderRepo(mongoDB.Conn)

	grpcClients, err := gclients.NewClients(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC clients: %w", err)
	}
	inventoryClient := gclients.NewInventoryClient(grpcClients.Inventory)

	producer, err := kafka.NewKafkaProducer(cfg.Brokers, "order.created")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kafka producer: %w", err)
	}

	orderUsecase := usecase.NewOrder(aiRepo, orderRepo, inventoryClient, producer)

	//httpServer := httpRepo.New(cfg.Server, orderUsecase)
	grpcServer := grpcAPI.New(cfg.Server, orderUsecase)

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
	app.grpcClients.Close()
	if err := app.kafkaProd.Close(); err != nil {
		log.Println("failed to close Kafka producer:", err)
	}
}
