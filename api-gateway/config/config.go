package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"time"
)

type Config struct {
	HTTPServer HTTPServer
	Services   Microservices
}

type HTTPServer struct {
	Port           int           `env:"HTTP_PORT,required"`
	ReadTimeout    time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"30s"`
	WriteTimeout   time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout    time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	MaxHeaderBytes int           `env:"HTTP_MAX_HEADER_BYTES" envDefault:"1048576"`
	Mode           string        `env:"GIN_MODE" envDefault:"release"`
}

type Microservices struct {
	UserService      ServiceConfig `envPrefix:"USER_SERVICE_"`
	InventoryService ServiceConfig `envPrefix:"INVENTORY_SERVICE_"`
	OrderService     ServiceConfig `envPrefix:"ORDER_SERVICE_"`
}

type ServiceConfig struct {
	Host string `env:"HOST,required"`
	Port int    `env:"PORT,required"`
}

func New() (*Config, error) {
	if err := godotenv.Load("local.env"); err != nil {
		log.Printf("Error loading local.env file: %v", err)
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
