package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port    int
	Timeout time.Duration
	Debug   bool
}

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

type AuthConfig struct {
    JWTSecret   string
    TokenExpiry time.Duration
    AdminSecret string 
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
}

func LoadConfig(path string) (*Config, error) {
    cfg := &Config{
        Server: ServerConfig{
            Port:  8080,
            Debug: false,
        },
        Database: DatabaseConfig{
            Host:            "localhost",
            Port:            5432,
            User:            "postgres",
            Password:        "password123",
            DBName:          "testordentdb",
            SSLMode:         "disable",
            MaxOpenConns:    20,
            MaxIdleConns:    5,
            ConnMaxLifetime: time.Hour,
        },
        Auth: AuthConfig{
            JWTSecret:   "super-secure-jwt-secret-key-123",
            TokenExpiry: 24 * time.Hour,
        },
    }

    file, err := os.Open(path)
    if err != nil {
        return cfg, fmt.Errorf("failed to open config file: %w", err)
    }
    defer file.Close()

    decoder := yaml.NewDecoder(file)
    if err := decoder.Decode(cfg); err != nil {
        return cfg, fmt.Errorf("failed to decode config file: %w", err)
    }

    if cfg.Auth.JWTSecret == "" {
        return cfg, fmt.Errorf("JWT secret cannot be empty")
    }

    return cfg, nil
}

