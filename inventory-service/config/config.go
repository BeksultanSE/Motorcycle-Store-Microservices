package config

import (
	"github.com/BeksultanSE/Assignment1-inventory/pkg/mongo"
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"time"
)

type (
	Config struct {
		Mongo   mongo.Config
		Server  Server
		Redis   Redis
		Cache   Cache
		Brokers []string `env:"BROKERS"`
		Version string   `env:"VERSION"`
	}

	Server struct {
		HTTPServer HTTPServer
		GRPCServer GRPCServer
	}

	HTTPServer struct {
		Port           int           `env:"HTTP_PORT,required"`
		ReadTimeout    time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"30s"`
		WriteTimeout   time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
		IdleTimeout    time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
		MaxHeaderBytes int           `env:"HTTP_MAX_HEADER_BYTES" envDefault:"1048576"` // 1 MB
		Mode           string        `env:"GIN_MODE" envDefault:"release"`              // Can be: release, debug, test
	}

	GRPCServer struct {
		Port    int           `env:"GRPC_PORT,required"`
		Timeout time.Duration `env:"GRPC_TIMEOUT" envDefault:"30s"`
	}

	// Redis configuration for main application
	Redis struct {
		Host         string        `env:"REDIS_HOSTS,notEmpty" envSeparator:","`
		Password     string        `env:"REDIS_PASSWORD"`
		TLSEnable    bool          `env:"REDIS_TLS_ENABLE" envDefault:"true"`
		DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" envDefault:"60s"`
		WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT" envDefault:"60s"`
		ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" envDefault:"30s"`
	}

	//Cache configuration
	Cache struct {
		TTL time.Duration `env:"CACHE_TTL" envDefault:"24h"`
		//refresh config if needed
	}
)

func New() (*Config, error) {
	//Loading local .env file for private configuration
	if err := godotenv.Load("local.env"); err != nil {
		log.Printf("Error loading local.env file")
	}

	var cfg Config
	err := env.Parse(&cfg)

	return &cfg, err
}
