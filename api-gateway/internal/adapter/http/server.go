package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/BeksultanSE/Assignment1-api-gateway/config"
	"github.com/BeksultanSE/Assignment1-api-gateway/internal/adapter/http/handler"
	"github.com/BeksultanSE/Assignment1-api-gateway/internal/adapter/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const serverIPAddress = "0.0.0.0:%d"

type Server struct {
	httpServer *gin.Engine
	cfg        config.HTTPServer

	address string
	handler *handler.Handler
}

func NewServer(cfg config.Config, handler *handler.Handler) *Server {
	gin.SetMode(cfg.HTTPServer.Mode)

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	api := &Server{
		httpServer: r,
		cfg:        cfg.HTTPServer,
		address:    fmt.Sprintf(serverIPAddress, cfg.HTTPServer.Port),
		handler:    handler,
	}

	api.setupRoutes()

	return api
}

func (s *Server) setupRoutes() {
	//exporting metrics
	middleware.PrometheusInit()
	s.httpServer.GET("/metrics", gin.WrapH(promhttp.Handler()))
	s.httpServer.Use(middleware.TrackMetrics())

	//api routes setup
	v1 := s.httpServer.Group("/api/v1")

	v1.POST("/users/register", s.handler.RegisterUser)
	v1.GET("/users/profile", middleware.AuthMiddleware(s.handler.Clients.User), s.handler.GetUserProfile)

	v1.GET("/products", s.handler.ListProducts) // Public endpoint

	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware(s.handler.Clients.User))
	{
		protected.POST("/products", s.handler.CreateProduct)
		protected.GET("/products/:id", s.handler.GetProduct)
		protected.PUT("/products/:id", s.handler.UpdateProduct)
		protected.DELETE("/products/:id", s.handler.DeleteProduct)

		protected.POST("/orders", s.handler.CreateOrder)
		protected.GET("/orders", s.handler.GetOrders)
		protected.GET("/orders/:id", s.handler.GetOrder)
		protected.PUT("/orders/:id", s.handler.UpdateOrder)
	}
	s.httpServer.NoRoute(func(c *gin.Context) {
		log.Printf("No route matched: %s %s", c.Request.Method, c.Request.URL.String())
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
	})
}

func (s *Server) Run(errCh chan<- error) {
	go func() {
		log.Printf("HTTP server running on: %v", s.address)

		if err := s.httpServer.Run(s.address); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to run HTTP server: %w", err)
			return
		}
	}()
}

func (s *Server) Stop() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Println("Shutdown signal received", "signal:", sig.String())

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("HTTP server shutting down gracefully")

	log.Println("HTTP server stopped successfully")

	return nil
}
