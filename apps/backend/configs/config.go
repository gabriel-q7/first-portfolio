package configs

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	AI          AIConfig
	ExternalAPIs ExternalAPIsConfig
	Auth        AuthConfig
	RateLimit   RateLimitConfig
	Observability ObservabilityConfig
	Environment string
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	RequestMaxBytes int64
}

type DatabaseConfig struct {
	Postgres PostgresConfig
	Redis    RedisConfig
}

type PostgresConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type AIConfig struct {
	Provider   string
	APIKey     string
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
	RateLimit  float64
}

type ExternalAPIsConfig struct {
	Timeout    time.Duration
	MaxRetries int
}

type AuthConfig struct {
	APIKeys   []string
	JWTSecret string
}

type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
}

type ObservabilityConfig struct {
	LogLevel        string
	MetricsEnabled  bool
	TracingEnabled  bool
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			RequestMaxBytes: getInt64("SERVER_REQUEST_MAX_BYTES", 1<<20), // 1 MB
		},
		Database: DatabaseConfig{
			Postgres: PostgresConfig{
				DSN:             getEnv("POSTGRES_DSN", ""),
				MaxOpenConns:    getInt("POSTGRES_MAX_OPEN_CONNS", 25),
				MaxIdleConns:    getInt("POSTGRES_MAX_IDLE_CONNS", 5),
				ConnMaxLifetime: getDuration("POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
			},
			Redis: RedisConfig{
				Addr:     getEnv("REDIS_ADDR", ""),
				Password: getEnv("REDIS_PASSWORD", ""),
				DB:       getInt("REDIS_DB", 0),
				PoolSize: getInt("REDIS_POOL_SIZE", 10),
			},
		},
		AI: AIConfig{
			Provider:   getEnv("AI_PROVIDER", "openai"),
			APIKey:     getEnv("AI_API_KEY", ""),
			BaseURL:    getEnv("AI_BASE_URL", "https://api.openai.com/v1"),
			Timeout:    getDuration("AI_TIMEOUT", 30*time.Second),
			MaxRetries: getInt("AI_MAX_RETRIES", 3),
			RateLimit:  getFloat64("AI_RATE_LIMIT", 10),
		},
		ExternalAPIs: ExternalAPIsConfig{
			Timeout:    getDuration("EXTERNAL_API_TIMEOUT", 10*time.Second),
			MaxRetries: getInt("EXTERNAL_API_MAX_RETRIES", 3),
		},
		Auth: AuthConfig{
			APIKeys:   getStringSlice("AUTH_API_KEYS"),
			JWTSecret: getEnv("JWT_SECRET", ""),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: getFloat64("RATE_LIMIT_RPS", 100),
			BurstSize:         getInt("RATE_LIMIT_BURST", 200),
		},
		Observability: ObservabilityConfig{
			LogLevel:       getEnv("LOG_LEVEL", "info"),
			MetricsEnabled: getBool("METRICS_ENABLED", true),
			TracingEnabled: getBool("TRACING_ENABLED", false),
		},
		Environment: getEnv("APP_ENV", "development"),
	}
	return cfg
}

// IsProd returns true if environment is production.
func (c *Config) IsProd() bool {
	return c.Environment == "production" || c.Environment == "prod"
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getInt64(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func getFloat64(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func getBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getStringSlice(key string) []string {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	var result []string
	for _, part := range strings.Split(v, ",") {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
