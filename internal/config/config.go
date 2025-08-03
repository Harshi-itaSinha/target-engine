// Package config loads the configurations for different environment
package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
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
	TTL             time.Duration `yaml:"ttl"`
	CleanupInterval time.Duration `yaml:"cleanupInterval"`
	MaxSize         int           `yaml:"maxSize"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool
	Port    string
	Path    string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver           string        `yaml:"driver"`
	ConnectionString string        `yaml:"uri"`
	MaxOpenConns     int           `yaml:"maxOpenConns"`
	MaxIdleConns     int           `yaml:"maxIdleConns"`
	ConnMaxLifetime  time.Duration `yaml:"connMaxLifetime"`
	DatabaseName     string        `yaml:"name"`
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
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev" // fallback to dev if not set
	}

	getConfigPath:= getConfigPath("config.dev.yml")
	data, err := ioutil.ReadFile(getConfigPath)
	if err != nil {
		log.Fatalf("failed to read config file '%s': %v",getConfigPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}
	return &cfg
}

func getConfigPath(filename string) string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println("Working directory:", wd)
	return filepath.Join(wd, "internal", "config", filename)
}

func GetEnv(key string) string {
	return os.Getenv(key)
}


