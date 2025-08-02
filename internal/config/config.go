// Package config loads the configurations for different environment
package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/stretchr/testify/assert/yaml"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	Cache     CacheConfig
	Metrics   MetricsConfig
	Database  DatabaseConfig
	RateLimit RateLimitConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	TTL             time.Duration
	CleanupInterval time.Duration
	MaxSize         int
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool
	Port    string
	Path    string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver           string
	ConnectionString string
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled    bool
	RPS        int
	BurstSize  int
	WindowSize time.Duration
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev" // fallback to dev if not set
	}

	path := fmt.Sprintf("config/config.%s.yml", env)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config file '%s': %v", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	return &cfg
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv gets an integer environment variable with a default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getBoolEnv gets a boolean environment variable with a default value
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getDurationEnv gets a duration environment variable with a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
