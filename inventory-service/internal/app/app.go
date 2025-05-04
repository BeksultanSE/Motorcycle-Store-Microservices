package app

import (
	"context"
	"fmt"
	"github.com/BeksultanSE/Assignment1-inventory/config"
	grpcAPI "github.com/BeksultanSE/Assignment1-inventory/internal/adapter/grpc"
	"github.com/BeksultanSE/Assignment1-inventory/internal/adapter/kafka"
	"github.com/IBM/sarama"

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
	grpcServer    *grpcAPI.ServerAPI
	consumerGroup sarama.ConsumerGroup
	kafkaHandler  *kafka.Consumer
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

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, "inventory-consumer-group", kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}
	kafkaHandler := kafka.NewConsumer(pUsecase, "order.created")

	app := &App{
		//httpServer: httpServer,
		grpcServer:    grpcServer,
		consumerGroup: consumerGroup,
		kafkaHandler:  kafkaHandler,
	}

	return app, nil
}

func (app *App) Start() error {
	errCh := make(chan error)

	//app.httpServer.Run(errCh)
	app.grpcServer.Run(errCh)

	// Start Kafka consumer in background
	go func() {
		for {
			if err := app.consumerGroup.Consume(context.Background(), []string{app.kafkaHandler.Topic}, app.kafkaHandler); err != nil {
				log.Printf("Consumer error: %v", err)
			}
		}
	}()

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

	if err := app.consumerGroup.Close(); err != nil {
		log.Println("failed to close consumer group:", err)
	}
}
