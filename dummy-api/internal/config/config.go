package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config contains runtime configuration for the dummy service.
type Config struct {
	HTTPAddr           string `envconfig:"DUMMY_HTTP_ADDR" default:":8082"`
	GRPCAddr           string `envconfig:"DUMMY_GRPC_ADDR" default:":9082"`
	DatabaseURL        string `envconfig:"DUMMY_DATABASE_URL" required:"true"`
	IdentityGRPCTarget string `envconfig:"DUMMY_IDENTITY_GRPC_TARGET" default:"localhost:9081"`
}

// Load retrieves configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("load dummy config: %w", err)
	}
	return &cfg, nil
}
