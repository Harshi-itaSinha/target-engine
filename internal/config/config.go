package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Metrics  MetricsConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type MetricsConfig struct {
	Enabled bool
	Port    string
	Path    string
}

type DatabaseConfig struct {
	Driver           string
	ConnectionString string
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},

		Metrics: MetricsConfig{
			Enabled: getBoolEnv("METRICS_ENABLED", true),
			Port:    getEnv("METRICS_PORT", "9090"),
			Path:    getEnv("METRICS_PATH", "/metrics"),
		},
		Database: DatabaseConfig{
			Driver:           getEnv("DB_DRIVER", "memory"),
			ConnectionString: getEnv("DB_CONNECTION_STRING", ""),
			MaxOpenConns:     getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:     getIntEnv("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime:  getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
