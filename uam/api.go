package uam

import (
	"context"
	"errors"
	"github.com/Av1shay/di-demo/pkg/types"
	"time"
)

const cacheTTL = 10 * time.Minute

type ItemRepository interface {
	GetItemByName(ctx context.Context, name string) (types.Item, error)
	SaveItem(ctx context.Context, input types.ItemCreateInput) (types.Item, error)
	ListItems(ctx context.Context) ([]types.Item, error)
	DeleteItem(ctx context.Context, id string) error
}

type Cache interface {
	Get(ctx context.Context, key string, out any) error
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type Config struct {
	CacheEnabled bool
}

type API struct {
	cfg      Config
	itemRepo ItemRepository
	cache    Cache
}

func NewAPI(cfg Config, itemRepo ItemRepository, cache Cache) (*API, error) {
	if cfg.CacheEnabled && cache == nil {
		return nil, errors.New("cache not set")
	}
	return &API{
		cfg:      cfg,
		itemRepo: itemRepo,
		cache:    cache,
	}, nil
}
