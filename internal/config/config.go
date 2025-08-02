package config

import (
	"os"
	"strconv"
	"time"
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
	Driver          string
	ConnectionString string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
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
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Cache: CacheConfig{
			TTL:             getDurationEnv("CACHE_TTL", 5*time.Minute),
			CleanupInterval: getDurationEnv("CACHE_CLEANUP_INTERVAL", 10*time.Minute),
			MaxSize:         getIntEnv("CACHE_MAX_SIZE", 10000),
		},
		Metrics: MetricsConfig{
			Enabled: getBoolEnv("METRICS_ENABLED", true),
			Port:    getEnv("METRICS_PORT", "9090"),
			Path:    getEnv("METRICS_PATH", "/metrics"),
		},
		Database: DatabaseConfig{
			Driver:          getEnv("DB_DRIVER", "memory"),
			ConnectionString: getEnv("DB_CONNECTION_STRING", ""),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		RateLimit: RateLimitConfig{
			Enabled:    getBoolEnv("RATE_LIMIT_ENABLED", true),
			RPS:        getIntEnv("RATE_LIMIT_RPS", 1000),
			BurstSize:  getIntEnv("RATE_LIMIT_BURST", 2000),
			WindowSize: getDurationEnv("RATE_LIMIT_WINDOW", time.Minute),
		},
	}
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