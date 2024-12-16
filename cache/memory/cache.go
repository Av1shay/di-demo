package memory

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

type Cache struct {
	storage sync.Map
}

type cacheItem struct {
	value      []byte
	expiration time.Time
}

func NewCache() *Cache {
	return &Cache{}
}

func (c *Cache) Get(_ context.Context, key string, out any) error {
	if out == nil {
		return errors.New("empty output destination provided")
	}
	raw, ok := c.storage.Load(key)
	if !ok {
		return errors.New("key not found")
	}

	item := raw.(cacheItem)
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		c.storage.Delete(key)
		return errors.New("key expired")
	}

	return json.Unmarshal(item.value, out)
}

func (c *Cache) Set(_ context.Context, key string, val any, ttl time.Duration) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	item := cacheItem{value: data}
	if ttl > 0 {
		item.expiration = time.Now().Add(ttl)
	}
	c.storage.Store(key, item)
	return nil
}

func (c *Cache) Delete(_ context.Context, key string) error {
	c.storage.Delete(key)
	return nil
}
