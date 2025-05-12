package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BeksultanSE/Assignment1-inventory/internal/domain"
	"github.com/BeksultanSE/Assignment1-inventory/pkg/redis"
	goredis "github.com/redis/go-redis/v9"
	"time"
)

const (
	keyPrefix = "product:%d"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(client *redis.Client, ttl time.Duration) *RedisCache {
	return &RedisCache{
		client: client,
		ttl:    ttl,
	}
}

func (r *RedisCache) Get(ctx context.Context, productID uint64) (domain.Product, error) {
	data, err := r.client.Unwrap().Get(ctx, r.key(productID)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return domain.Product{}, nil // cache miss
		}
		return domain.Product{}, fmt.Errorf("failed to get product: %w", err)
	}

	var product domain.Product
	err = json.Unmarshal(data, &product)
	if err != nil {
		return domain.Product{}, fmt.Errorf("failed to unmarshal product: %w", err)
	}

	return product, nil
}

func (r *RedisCache) Set(ctx context.Context, product domain.Product) error {
	data, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product: %w", err)
	}

	return r.client.Unwrap().Set(ctx, r.key(product.ID), data, r.ttl).Err()
}

func (r *RedisCache) SetMany(ctx context.Context, products []domain.Product) error {
	pipe := r.client.Unwrap().Pipeline()
	for _, product := range products {
		data, err := json.Marshal(product)
		if err != nil {
			return fmt.Errorf("failed to marshal product: %w", err)
		}
		pipe.Set(ctx, r.key(product.ID), data, r.ttl)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set many products: %w", err)
	}
	return nil
}

func (r *RedisCache) Delete(ctx context.Context, productID uint64) error {
	return r.client.Unwrap().Del(ctx, r.key(productID)).Err()
}

// key cache method for constructing key
func (r *RedisCache) key(productID uint64) string {
	return fmt.Sprintf(keyPrefix, productID)
}
