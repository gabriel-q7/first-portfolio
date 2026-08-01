package configs

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	AI            AIConfig
	Auth          AuthConfig
	RateLimit     RateLimitConfig
	Observability ObservabilityConfig
	Environment   string
}

type ServerConfig struct {
	Port              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	RequestMaxBytes   int64
	MaxHeaderBytes    int
	TrustedProxyCIDR  string
}

type DatabaseConfig struct {
	SQLite SQLiteConfig
}

type SQLiteConfig struct {
	Path         string
	BusyTimeout  time.Duration
	MaxOpenConns int
}

type AIConfig struct {
	APIKey    string
	BaseURL   string
	RateLimit float64
}

type AuthConfig struct {
	APIKeys []string
}

type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	MaxVisitors       int
}

type ObservabilityConfig struct {
	LogLevel string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:              getEnv("SERVER_PORT", "8080"),
			ReadHeaderTimeout: getDuration("SERVER_READ_HEADER_TIMEOUT", 5*time.Second),
			ReadTimeout:       getDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:      getDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:       getDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout:   getDuration("SERVER_SHUTDOWN_TIMEOUT", 15*time.Second),
			RequestMaxBytes:   getInt64("SERVER_REQUEST_MAX_BYTES", 1<<20),
			MaxHeaderBytes:    getInt("SERVER_MAX_HEADER_BYTES", 16<<10),
			TrustedProxyCIDR:  getEnv("TRUSTED_PROXY_CIDR", ""),
		},
		Database: DatabaseConfig{
			SQLite: SQLiteConfig{
				Path:         getEnv("SQLITE_PATH", "/data/portfolio.db"),
				BusyTimeout:  getDuration("SQLITE_BUSY_TIMEOUT", 5*time.Second),
				MaxOpenConns: getInt("SQLITE_MAX_OPEN_CONNS", 2),
			},
		},
		AI: AIConfig{
			APIKey:    getEnv("AI_API_KEY", ""),
			BaseURL:   getEnv("AI_BASE_URL", "https://api.openai.com/v1"),
			RateLimit: getFloat64("AI_RATE_LIMIT", 2),
		},
		Auth: AuthConfig{
			APIKeys: getStringSlice("AUTH_API_KEYS"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: getFloat64("RATE_LIMIT_RPS", 20),
			BurstSize:         getInt("RATE_LIMIT_BURST", 40),
			MaxVisitors:       getInt("RATE_LIMIT_MAX_VISITORS", 2048),
		},
		Observability: ObservabilityConfig{
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		Environment: getEnv("APP_ENV", "production"),
	}
}

func (c *Config) IsProd() bool {
	return c.Environment == "production" || c.Environment == "prod"
}

func getEnv(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return def
	}
	return duration
}

func getInt(key string, def int) int {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return number
}

func getInt64(key string, def int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return def
	}
	return number
}

func getFloat64(key string, def float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return def
	}
	return number
}

func getStringSlice(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}
	result := make([]string, 0)
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
