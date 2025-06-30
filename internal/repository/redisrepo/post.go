package redisrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"blog/internal/config"
	"blog/internal/models"
)

var (
	KeyNotFound = errors.New("key not found")
)

type PostModel struct {
	models.Post
	Username string `json:"username"`
}

type RedisRepo struct {
	RDB *redis.Client
}

func NewRedisClient(cfg *config.Config) (*RedisRepo, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       0,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisRepo{RDB: rdb}, nil
}

func (rp *RedisRepo) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("(redis) failed to marshal: %w", err)
	}

	err = rp.RDB.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

func (rp *RedisRepo) Get(ctx context.Context, key string, dest any) error {
	data, err := rp.RDB.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return KeyNotFound
	} else if err != nil {
		return fmt.Errorf("redis get failed: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	return nil
}

func (rp *RedisRepo) Delete(ctx context.Context, key string) error {
	return rp.RDB.Del(ctx, key).Err()
}
