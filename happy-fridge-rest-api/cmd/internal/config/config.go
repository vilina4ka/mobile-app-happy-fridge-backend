package config

import (
	"fmt"
	"log"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the application.
type Config struct {
	BindAddr string `toml:"bind_addr" envconfig:"BIND_ADDR" default:":8080"`
	LogLevel string `toml:"log_level" envconfig:"LOG_LEVEL" default:"debug"`

	DatabaseURL       string        `toml:"database_url" envconfig:"DATABASE_URL" required:"true"`
	DBMaxOpenConns    int           `toml:"db_max_open_conns" envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	DBMaxIdleConns    int           `toml:"db_max_idle_conns" envconfig:"DB_MAX_IDLE_CONNS" default:"25"`
	DBConnMaxLifetime time.Duration `toml:"db_conn_max_lifetime" envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`

	JwtSecretKey      string        `toml:"jwt_secret_key" envconfig:"JWT_SECRET_KEY" required:"true"`
	JwtExpirationTime time.Duration `toml:"jwt_expiration_time" envconfig:"JWT_EXPIRATION_TIME" default:"24h"`
}

// LoadConfig loads configuration from file and environment variables.
func LoadConfig(configPath string) (*Config, error) {
	cfg := &Config{}

	if configPath != "" {
		if _, err := toml.DecodeFile(configPath, cfg); err != nil {
			log.Printf("Warning: Could not decode config file '%s': %v. Using defaults and environment variables.", configPath, err)
		}
	} else {
		log.Printf("Warning: No config file path provided. Using defaults and environment variables.")
	}

	err := envconfig.Process("APP", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
	}

	return cfg, nil
}
