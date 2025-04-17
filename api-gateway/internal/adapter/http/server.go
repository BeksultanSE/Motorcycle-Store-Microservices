package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/BeksultanSE/Assignment1-api-gateway/config"
	"github.com/BeksultanSE/Assignment1-api-gateway/internal/adapter/http/handler"
	"github.com/BeksultanSE/Assignment1-api-gateway/internal/adapter/http/middleware"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const serverIPAddress = "0.0.0.0:%d"

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg config.Config, handler *handler.Handler) *Server {
	gin.SetMode(cfg.HTTPServer.Mode)
	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")

	api.POST("/users/register", handler.RegisterUser)
	api.GET("/users/profile", middleware.AuthMiddleware(handler.Clients.User), handler.GetUserProfile)

	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(handler.Clients.User))
	{
		protected.POST("/products", handler.CreateProduct)
		protected.POST("/orders", handler.CreateOrder)
		protected.GET("/orders", handler.GetOrders)
	}

	r.NoRoute(func(c *gin.Context) {
		log.Printf("No route matched: %s %s", c.Request.Method, c.Request.URL.String())
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
	})

	httpServer := &http.Server{
		Addr:           fmt.Sprintf(serverIPAddress, cfg.HTTPServer.Port),
		Handler:        r,
		ReadTimeout:    cfg.HTTPServer.ReadTimeout,
		WriteTimeout:   cfg.HTTPServer.WriteTimeout,
		IdleTimeout:    cfg.HTTPServer.IdleTimeout,
		MaxHeaderBytes: cfg.HTTPServer.MaxHeaderBytes,
	}

	return &Server{httpServer: httpServer}
}

func (s *Server) Run(errCh chan<- error) {
	go func() {
		log.Printf("HTTP server running on: %v", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to run HTTP server: %w", err)
		}
	}()
}

func (s *Server) Stop() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Println("Shutdown signal received", "signal:", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("HTTP server shutting down gracefully")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Println("HTTP server shutdown error:", err)
		return err
	}
	log.Println("HTTP server stopped successfully")

	return nil
}
