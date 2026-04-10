package nosql

import (
	"context"
	"fmt"
	"time"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/repository"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// RedisConfig holds Redis connection parameters.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type redisCacheRepo struct {
	client *redis.Client
	logger logger.Logger
}

// New returns a Redis-backed CacheRepository.
func New(client *redis.Client, log logger.Logger) repository.CacheRepository {
	return &redisCacheRepo{client: client, logger: log}
}

// Connect creates a Redis client and verifies connectivity.
func Connect(cfg RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return client, nil
}

func (r *redisCacheRepo) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache Get %q: %w", key, err)
	}
	return val, nil
}

func (r *redisCacheRepo) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("cache Set %q: %w", key, err)
	}
	return nil
}

func (r *redisCacheRepo) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache Delete %q: %w", key, err)
	}
	return nil
}

func (r *redisCacheRepo) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("cache Exists %q: %w", key, err)
	}
	return n > 0, nil
}

func (r *redisCacheRepo) FlushPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("cache FlushPattern scan %q: %w", pattern, err)
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("cache FlushPattern del: %w", err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
