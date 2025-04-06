package config

import (
	"github.com/BeksultanSE/Assignment1-inventory/pkg/mongo"
	"github.com/caarlos0/env/v10"
	_ "github.com/joho/godotenv/autoload"
	"time"
)

type (
	Config struct {
		Mongo  mongo.Config
		Server Server

		Version string `env:"VERSION"`
	}

	Server struct {
		HTTPServer HTTPServer
	}

	HTTPServer struct {
		Port           int           `env:"HTTP_PORT,required"`
		ReadTimeout    time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"30s"`
		WriteTimeout   time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
		IdleTimeout    time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
		MaxHeaderBytes int           `env:"HTTP_MAX_HEADER_BYTES" envDefault:"1048576"` // 1 MB
		Mode           string        `env:"GIN_MODE" envDefault:"release"`              // Can be: release, debug, test
	}
)

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)

	return &cfg, err
}
