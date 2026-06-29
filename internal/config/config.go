package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Log       LogConfig
	Migration MigrationConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port            int           `mapstructure:"SERVER_PORT"`
	ShutdownTimeout time.Duration `mapstructure:"SHUTDOWN_TIMEOUT"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	URL string `mapstructure:"DATABASE_URL"`
}

// JWTConfig holds JWT token configuration.
type JWTConfig struct {
	Secret             string        `mapstructure:"JWT_SECRET"`
	AccessTokenExpiry  time.Duration `mapstructure:"ACCESS_TOKEN_EXPIRY"`
	RefreshTokenExpiry time.Duration `mapstructure:"REFRESH_TOKEN_EXPIRY"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `mapstructure:"LOG_LEVEL"`
}

// MigrationConfig holds database migration configuration.
type MigrationConfig struct {
	AutoMigrate bool `mapstructure:"AUTO_MIGRATE"`
}

// Load reads configuration from environment variables and optionally from a .env file.
// Environment variables take precedence over .env file values.
// Returns a descriptive error indicating the field and violated constraint if validation fails.
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("SERVER_PORT", 8080)
	v.SetDefault("SHUTDOWN_TIMEOUT", "30s")
	v.SetDefault("ACCESS_TOKEN_EXPIRY", "15m")
	v.SetDefault("REFRESH_TOKEN_EXPIRY", "168h")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("AUTO_MIGRATE", false)

	// Read from .env file if present (optional)
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig() // ignore error if .env doesn't exist

	// Environment variables take precedence
	v.AutomaticEnv()

	cfg := &Config{
		Server: ServerConfig{
			Port:            v.GetInt("SERVER_PORT"),
			ShutdownTimeout: v.GetDuration("SHUTDOWN_TIMEOUT"),
		},
		Database: DatabaseConfig{
			URL: v.GetString("DATABASE_URL"),
		},
		JWT: JWTConfig{
			Secret:             v.GetString("JWT_SECRET"),
			AccessTokenExpiry:  v.GetDuration("ACCESS_TOKEN_EXPIRY"),
			RefreshTokenExpiry: v.GetDuration("REFRESH_TOKEN_EXPIRY"),
		},
		Log: LogConfig{
			Level: v.GetString("LOG_LEVEL"),
		},
		Migration: MigrationConfig{
			AutoMigrate: v.GetBool("AUTO_MIGRATE"),
		},
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks all configuration constraints and returns a descriptive error
// indicating the field name and the violated constraint.
func validate(cfg *Config) error {
	// Server port must be between 1024 and 65535
	if cfg.Server.Port < 1024 || cfg.Server.Port > 65535 {
		return fmt.Errorf("config: SERVER_PORT must be between 1024 and 65535, got %d", cfg.Server.Port)
	}

	// Shutdown timeout must be positive
	if cfg.Server.ShutdownTimeout <= 0 {
		return fmt.Errorf("config: SHUTDOWN_TIMEOUT must be a positive duration, got %v", cfg.Server.ShutdownTimeout)
	}

	// DATABASE_URL is required
	if strings.TrimSpace(cfg.Database.URL) == "" {
		return fmt.Errorf("config: DATABASE_URL is required and cannot be empty")
	}

	// JWT secret must be at least 32 characters
	if len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("config: JWT_SECRET must be at least 32 characters, got %d", len(cfg.JWT.Secret))
	}

	// Access token expiry must be positive
	if cfg.JWT.AccessTokenExpiry <= 0 {
		return fmt.Errorf("config: ACCESS_TOKEN_EXPIRY must be a positive duration, got %v", cfg.JWT.AccessTokenExpiry)
	}

	// Refresh token expiry must be positive
	if cfg.JWT.RefreshTokenExpiry <= 0 {
		return fmt.Errorf("config: REFRESH_TOKEN_EXPIRY must be a positive duration, got %v", cfg.JWT.RefreshTokenExpiry)
	}

	// Log level must be one of the allowed values
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	level := strings.ToLower(cfg.Log.Level)
	if !validLevels[level] {
		return fmt.Errorf("config: LOG_LEVEL must be one of [debug, info, warn, error], got %q", cfg.Log.Level)
	}

	return nil
}
