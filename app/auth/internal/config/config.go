package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Limen  LimenConfig  `yaml:"limen"`
	Data   DataConfig   `yaml:"data"`
}

type ServerConfig struct {
	HTTP HTTPServerConfig `yaml:"http"`
}

type HTTPServerConfig struct {
	Addr string `yaml:"addr"`
}

type LimenConfig struct {
	BaseURL        string   `yaml:"base_url"`
	Secret         string   `yaml:"secret"`
	TrustedOrigins []string `yaml:"trusted_origins"`
	CookieDomain   string   `yaml:"cookie_domain"`
	FrontendURL    string   `yaml:"frontend_url"`
}

type DataConfig struct {
	Database DatabaseConfig `yaml:"database"`
	NATS     NATSConfig     `yaml:"nats"`
}

type DatabaseConfig struct {
	Source string `yaml:"source"`
}

type NATSConfig struct {
	URL string `yaml:"url"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if s := os.Getenv("LIMEN_SECRET"); s != "" {
		cfg.Limen.Secret = s
	}
	if u := os.Getenv("LIMEN_BASE_URL"); u != "" {
		cfg.Limen.BaseURL = u
	}
	if u := os.Getenv("FRONTEND_URL"); u != "" {
		cfg.Limen.FrontendURL = u
	}
	if d := os.Getenv("DATABASE_URL"); d != "" {
		cfg.Data.Database.Source = d
	}
	if n := os.Getenv("NATS_URL"); n != "" {
		cfg.Data.NATS.URL = n
	}
	if cfg.Server.HTTP.Addr == "" {
		cfg.Server.HTTP.Addr = "0.0.0.0:8080"
	}
	if cfg.Limen.FrontendURL == "" {
		cfg.Limen.FrontendURL = "http://localhost:3000"
	}
	return &cfg, nil
}
