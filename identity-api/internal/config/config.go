package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config holds runtime configuration loaded via environment variables.
type Config struct {
	HTTPAddr    string `envconfig:"IDENTITY_HTTP_ADDR" default:":8081"`
	GRPCAddr    string `envconfig:"IDENTITY_GRPC_ADDR" default:":9081"`
	DatabaseURL string `envconfig:"IDENTITY_DATABASE_URL" required:"true"`
	JWTSecret   string `envconfig:"IDENTITY_JWT_SECRET" default:"dev-secret"`
}

// Load reads environment variables into Config.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("load identity config: %w", err)
	}
	return &cfg, nil
}
