package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"toll_plaza/internal/domain"

	"github.com/redis/go-redis/v9"
)

type CacheService interface {
	GetRoute(ctx context.Context, sourcePincode, destPincode string) (*domain.TollResponse, bool)
	SetRoute(ctx context.Context, sourcePincode, destPincode string, resp *domain.TollResponse, ttl time.Duration)
	Ping(ctx context.Context) error
	Close() error
}

type redisCache struct {
	client *redis.Client
}

func NewRedisCache(redisURL string) CacheService {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("Warning: Failed to parse REDIS_URL (%s), falling back to localhost:6379: %v", redisURL, err)
		opt = &redis.Options{
			Addr: "localhost:6379",
		}
	}

	client := redis.NewClient(opt)
	return &redisCache{client: client}
}

func (r *redisCache) key(source, dest string) string {
	return fmt.Sprintf("toll:route:%s:%s", source, dest)
}

func (r *redisCache) GetRoute(ctx context.Context, sourcePincode, destPincode string) (*domain.TollResponse, bool) {
	if r.client == nil {
		return nil, false
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	val, err := r.client.Get(ctxTimeout, r.key(sourcePincode, destPincode)).Result()
	if err != nil {
		if err != redis.Nil {
			log.Printf("Redis GET error: %v", err)
		}
		return nil, false
	}

	var resp domain.TollResponse
	if err := json.Unmarshal([]byte(val), &resp); err != nil {
		log.Printf("Redis unmarshal error: %v", err)
		return nil, false
	}

	return &resp, true
}

func (r *redisCache) SetRoute(ctx context.Context, sourcePincode, destPincode string, resp *domain.TollResponse, ttl time.Duration) {
	if r.client == nil || resp == nil {
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Redis marshal error: %v", err)
		return
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := r.client.Set(ctxTimeout, r.key(sourcePincode, destPincode), string(data), ttl).Err(); err != nil {
		log.Printf("Redis SET error: %v", err)
	}
}

func (r *redisCache) Ping(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("redis client is nil")
	}
	return r.client.Ping(ctx).Err()
}

func (r *redisCache) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
