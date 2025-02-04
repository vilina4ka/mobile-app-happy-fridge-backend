package apiserver

import "happy-fridge-rest-api/cmd/internal/app/store"

// Config ...
type Config struct {
	BinAddr  string `toml:"bin_addr"`
	LogLevel string `toml:"log_level"`
	Store    *store.Config
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BinAddr:  ":8080",
		LogLevel: "debug",
		Store:    store.NewConfig(),
	}
}
