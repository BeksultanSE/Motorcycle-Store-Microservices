package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/BeksultanSE/Assignment1-order/config"
	"github.com/BeksultanSE/Assignment1-order/internal/adapter/http/handler"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const serverIPAddress = "0.0.0.0:%d" // Changed to 0.0.0.0 for external access

type API struct {
	server *gin.Engine
	cfg    config.HTTPServer

	address      string
	orderHandler *handler.OrderHandler // Changed to OrderHandler
}

func New(cfg config.Server, useCase handler.OrderUsecase) *API {
	// Setting the Gin mode
	gin.SetMode(cfg.HTTPServer.Mode)

	// Creating a new Gin Engine
	server := gin.New()

	// Applying middleware
	server.Use(gin.Recovery())

	// Binding orders
	orderHandler := handler.NewOrderHandler(useCase) // Changed to OrderHandler

	api := &API{
		server:       server,
		cfg:          cfg.HTTPServer,
		address:      fmt.Sprintf(serverIPAddress, cfg.HTTPServer.Port),
		orderHandler: orderHandler, // Changed to OrderHandler
	}

	api.setupRoutes()

	return api
}

func (api *API) setupRoutes() {
	v1 := api.server.Group("api/v1")
	{
		orders := v1.Group("/orders")
		{
			orders.POST("/", api.orderHandler.Create)     // Create a new order
			orders.GET("/:id", api.orderHandler.GetByID)  // Retrieve order details by ID
			orders.PATCH("/:id", api.orderHandler.Update) // Update order status
			orders.GET("", api.orderHandler.GetAll)       // List all orders for a user
		}
	}
}

func (api *API) Run(errCh chan<- error) {
	go func() {
		log.Printf("HTTP server running on: %v", api.address)

		// No need to reinitialize `api.server` here. Just run it directly.
		if err := api.server.Run(api.address); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to run HTTP server: %w", err)
			return
		}
	}()
}

func (a *API) Stop() error {
	// Setting up the signal channel to catch termination signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Blocking until a signal is received
	sig := <-quit
	log.Println("Shutdown signal received", "signal:", sig.String())

	// Creating a context with timeout for graceful shutdown
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("HTTP server shutting down gracefully")

	// Note: You can use `Shutdown` if you use `http.Server` instead of `gin.Engine`.
	log.Println("HTTP server stopped successfully")

	return nil
}
