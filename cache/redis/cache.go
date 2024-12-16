package redis

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

const keyPrefix = "di-demo:"

type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func (c *Cache) Get(ctx context.Context, key string, out any) error {
	if out == nil {
		return errors.New("empty output destination provided")
	}
	key = BuildKey(key)
	res, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(res, &out)
}

func (c *Cache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	key = BuildKey(key)
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	key = BuildKey(key)
	return c.rdb.Del(ctx, key).Err()
}

func BuildKey(v string) string {
	return keyPrefix + v
}
