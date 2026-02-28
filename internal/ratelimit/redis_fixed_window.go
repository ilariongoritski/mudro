package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisFixedWindow struct {
	client *redis.Client
	prefix string
}

func NewRedisFixedWindow(addr, password string, db int) *RedisFixedWindow {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisFixedWindow{
		client: c,
		prefix: "mudro:rl",
	}
}

func (l *RedisFixedWindow) Close() error {
	if l == nil || l.client == nil {
		return nil
	}
	return l.client.Close()
}

func (l *RedisFixedWindow) Ping(ctx context.Context) error {
	return l.client.Ping(ctx).Err()
}

func (l *RedisFixedWindow) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if limit <= 0 {
		return false, nil
	}
	if window <= 0 {
		window = time.Second
	}

	slot := time.Now().UTC().UnixNano() / window.Nanoseconds()
	redisKey := fmt.Sprintf("%s:%s:%d", l.prefix, key, slot)
	pipe := l.client.TxPipeline()
	incr := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, window+2*time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	n, err := incr.Result()
	if err != nil {
		return false, err
	}
	return n <= int64(limit), nil
}
