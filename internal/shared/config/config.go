package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server" json:"server"`
	Database DatabaseConfig `yaml:"database" json:"database"`
	Issuer   IssuerConfig   `yaml:"issuer" json:"issuer"`
	Rotation RotationConfig `yaml:"rotation" json:"rotation"`
	Webhook  WebhookConfig  `yaml:"webhook" json:"webhook"`
}

type ServerConfig struct {
	Address         string        `yaml:"address" json:"address"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	RateLimitRPS    int           `yaml:"rate_limit_rps" json:"rate_limit_rps"`
	RateLimitBurst  int           `yaml:"rate_limit_burst" json:"rate_limit_burst"`
}

type DatabaseConfig struct {
	Driver          string `yaml:"driver" json:"driver"`
	DSN             string `yaml:"dsn" json:"dsn"`
	MaxOpenConns    int    `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
}

type IssuerConfig struct {
	DefaultLifetimeDays int    `yaml:"default_lifetime_days" json:"default_lifetime_days"`
	CommonName          string `yaml:"common_name" json:"common_name"`
}

type RotationConfig struct {
	DefaultAdvanceDays int `yaml:"default_advance_days" json:"default_advance_days"`
}

type WebhookConfig struct {
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
}

func Load(path string) (Config, error) {
	cfg := defaults()
	if path == "" {
		return applyEnv(cfg), nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return applyEnv(cfg), nil
}

func defaults() Config {
	return Config{
		Server: ServerConfig{
			Address:         ":8080",
			ShutdownTimeout: 15 * time.Second,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    15 * time.Second,
			IdleTimeout:     60 * time.Second,
			RateLimitRPS:    100,
			RateLimitBurst:  200,
		},
		Database: DatabaseConfig{
			Driver:          "sqlite",
			DSN:             "certpilot.db",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: "30m",
		},
		Issuer: IssuerConfig{
			DefaultLifetimeDays: 90,
			CommonName:          "CertPilot Local Issuer",
		},
		Rotation: RotationConfig{DefaultAdvanceDays: 30},
		Webhook:  WebhookConfig{Timeout: 5 * time.Second},
	}
}

func applyEnv(cfg Config) Config {
	cfg.Server.Address = envString("CERTPILOT_SERVER_ADDRESS", cfg.Server.Address)
	cfg.Server.ShutdownTimeout = envDuration("CERTPILOT_SERVER_SHUTDOWN_TIMEOUT", cfg.Server.ShutdownTimeout)
	cfg.Server.ReadTimeout = envDuration("CERTPILOT_SERVER_READ_TIMEOUT", cfg.Server.ReadTimeout)
	cfg.Server.WriteTimeout = envDuration("CERTPILOT_SERVER_WRITE_TIMEOUT", cfg.Server.WriteTimeout)
	cfg.Server.IdleTimeout = envDuration("CERTPILOT_SERVER_IDLE_TIMEOUT", cfg.Server.IdleTimeout)
	cfg.Server.RateLimitRPS = envInt("CERTPILOT_SERVER_RATE_LIMIT_RPS", cfg.Server.RateLimitRPS)
	cfg.Server.RateLimitBurst = envInt("CERTPILOT_SERVER_RATE_LIMIT_BURST", cfg.Server.RateLimitBurst)
	cfg.Database.Driver = envString("CERTPILOT_DATABASE_DRIVER", cfg.Database.Driver)
	cfg.Database.DSN = envString("CERTPILOT_DATABASE_DSN", cfg.Database.DSN)
	cfg.Database.MaxOpenConns = envInt("CERTPILOT_DATABASE_MAX_OPEN_CONNS", cfg.Database.MaxOpenConns)
	cfg.Database.MaxIdleConns = envInt("CERTPILOT_DATABASE_MAX_IDLE_CONNS", cfg.Database.MaxIdleConns)
	cfg.Database.ConnMaxLifetime = envString("CERTPILOT_DATABASE_CONN_MAX_LIFETIME", cfg.Database.ConnMaxLifetime)
	cfg.Issuer.DefaultLifetimeDays = envInt("CERTPILOT_ISSUER_DEFAULT_LIFETIME_DAYS", cfg.Issuer.DefaultLifetimeDays)
	cfg.Issuer.CommonName = envString("CERTPILOT_ISSUER_COMMON_NAME", cfg.Issuer.CommonName)
	cfg.Rotation.DefaultAdvanceDays = envInt("CERTPILOT_ROTATION_DEFAULT_ADVANCE_DAYS", cfg.Rotation.DefaultAdvanceDays)
	cfg.Webhook.Timeout = envDuration("CERTPILOT_WEBHOOK_TIMEOUT", cfg.Webhook.Timeout)
	return cfg
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Server.Address) == "" {
		return errors.New("server.address is required")
	}
	if c.Database.Driver != "sqlite" && c.Database.Driver != "postgres" {
		return fmt.Errorf("unsupported database driver %q", c.Database.Driver)
	}
	if strings.TrimSpace(c.Database.DSN) == "" {
		return errors.New("database.dsn is required")
	}
	if c.Issuer.DefaultLifetimeDays <= 0 {
		return errors.New("issuer.default_lifetime_days must be positive")
	}
	return nil
}

func envString(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			return parsed
		}
	}
	return fallback
}
