// Package config provides application configuration management.
package config

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// DatabaseType represents the type of database driver
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgresql"
	MySQL      DatabaseType = "mysql"
	SQLite     DatabaseType = "sqlite"
	SQLServer  DatabaseType = "sqlserver"
)

// Config holds all application configuration
type Config struct {
	App       AppConfig
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	GRPC      GRPCConfig
	Jaeger    JaegerConfig
	RateLimit RateLimitConfig
}

type AppConfig struct {
	Name        string
	Version     string
	Environment string
	LogLevel    string
}

type ServerConfig struct {
	Host         string
	Port         string
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Type            DatabaseType
	Host            string
	Port            int
	Username        string
	Password        string
	Database        string
	Schema          string
	SSLMode         string
	SSLCert         string
	SSLKey          string
	SSLRootCert     string
	CustomURL       string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type RedisConfig struct {
	Enabled      bool
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpire  time.Duration
	RefreshTokenExpire time.Duration
}

type GRPCConfig struct {
	Enabled bool
	Host    string
	Port    string
}

type JaegerConfig struct {
	Enabled     bool
	Endpoint    string
	ServiceName string
}

type RateLimitConfig struct {
	Enabled bool
	RPS     float64
	Burst   int
}

var (
	instance *Config
	once     sync.Once
)

// GetInstance returns the singleton config instance
func GetInstance() *Config {
	once.Do(func() {
		_ = godotenv.Load()
		instance = load()
	})
	return instance
}

func load() *Config {
	return &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "cms-backend"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
			Environment: getEnv("APP_ENV", "development"),
			LogLevel:    getEnv("APP_LOG_LEVEL", "info"),
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnv("SERVER_PORT", "8080"),
			Mode:         getEnv("SERVER_MODE", "debug"),
			ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			Type:            DatabaseType(getEnv("DB_TYPE", "sqlite")),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getInt("DB_PORT", 5432),
			Username:        getEnv("DB_USERNAME", ""),
			Password:        getEnv("DB_PASSWORD", ""),
			Database:        getEnv("DB_DATABASE", "cms_backend.db"),
			Schema:          getEnv("DB_SCHEMA", "public"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			SSLCert:         getEnv("DB_SSL_CERT", ""),
			SSLKey:          getEnv("DB_SSL_KEY", ""),
			SSLRootCert:     getEnv("DB_SSL_ROOT_CERT", ""),
			CustomURL:       getEnv("DB_CUSTOM_URL", ""),
			MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 300*time.Second),
			ConnMaxIdleTime: getDuration("DB_CONN_MAX_IDLE_TIME", 60*time.Second),
		},
		Redis: RedisConfig{
			Enabled:      getBool("REDIS_ENABLED", false),
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getInt("REDIS_DB", 0),
			PoolSize:     getInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getInt("REDIS_MIN_IDLE_CONNS", 2),
			MaxRetries:   getInt("REDIS_MAX_RETRIES", 3),
			DialTimeout:  getDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "default-secret-change-in-production"),
			AccessTokenExpire:  getDuration("JWT_ACCESS_TOKEN_EXPIRE", 15*time.Minute),
			RefreshTokenExpire: getDuration("JWT_REFRESH_TOKEN_EXPIRE", 7*24*time.Hour),
		},
		GRPC: GRPCConfig{
			Enabled: getBool("GRPC_ENABLED", false),
			Host:    getEnv("GRPC_HOST", "0.0.0.0"),
			Port:    getEnv("GRPC_PORT", "9090"),
		},
		Jaeger: JaegerConfig{
			Enabled:     getBool("JAEGER_ENABLED", false),
			Endpoint:    getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
			ServiceName: getEnv("JAEGER_SERVICE_NAME", "cms-backend"),
		},
		RateLimit: RateLimitConfig{
			Enabled: getBool("RATE_LIMIT_ENABLED", true),
			RPS:     getFloat64("RATE_LIMIT_RPS", 100),
			Burst:   getInt("RATE_LIMIT_BURST", 200),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func getBool(key string, defaultVal bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultVal
}

func getFloat64(key string, defaultVal float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		// Try as seconds integer
		if s, err := strconv.Atoi(v); err == nil {
			return time.Duration(s) * time.Second
		}
	}
	return defaultVal
}
