package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	CORS     CORSConfig
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port    int
	Timeout time.Duration
	Debug   bool
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	JWTSecret   string
	TokenExpiry time.Duration
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
}

func LoadConfig(path string) (*Config, error) {
    // Buat konfigurasi default
    cfg := &Config{
        Server: ServerConfig{
            Port:  8080,
            Debug: false,
        },
        Database: DatabaseConfig{
            Host:            "localhost",
            Port:            5432,
            User:            "postgres",
            Password:        "postgres",
            DBName:          "ecommerce",
            SSLMode:         "disable",
            MaxOpenConns:    20,
            MaxIdleConns:    5,
            ConnMaxLifetime: time.Hour,
        },
        Auth: AuthConfig{
            JWTSecret:   "default-jwt-secret-key",  // Default value
            TokenExpiry: 24 * time.Hour,
        },
    }

    // Baca file konfigurasi
    file, err := os.Open(path)
    if err != nil {
        return cfg, fmt.Errorf("failed to open config file: %w", err)
    }
    defer file.Close()

    // Parse konfigurasi YAML
    decoder := yaml.NewDecoder(file)
    if err := decoder.Decode(cfg); err != nil {
        return cfg, fmt.Errorf("failed to decode config file: %w", err)
    }

    // Validasi konfigurasi
    if cfg.Auth.JWTSecret == "" {
        return cfg, fmt.Errorf("JWT secret cannot be empty")
    }

    return cfg, nil
}

