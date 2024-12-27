package uam

import (
	"context"
	"errors"
	"github.com/Av1shay/di-demo/pkg/types"
	"time"
)

const (
	itemCacheTTL = time.Hour
	listCacheTTL = time.Minute
)

type ItemRepository interface {
	GetItemByName(ctx context.Context, name, accountID string) (types.Item, error)
	SaveItem(ctx context.Context, input types.ItemCreateInput, accountID string) (types.Item, error)
	UpdateItem(ctx context.Context, input types.UpdateItemInput, accountID string) (types.Item, error)
	ListItems(ctx context.Context, input types.ListItemsInput, accountID string) ([]types.Item, error)
	DeleteItem(ctx context.Context, id, accountID string) error
}

type Repository interface {
	ItemRepository
	Ping(ctx context.Context) error
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
	cfg   Config
	repo  Repository
	cache Cache
}

func NewAPI(cfg Config, repo Repository, cache Cache) (*API, error) {
	if cfg.CacheEnabled && cache == nil {
		return nil, errors.New("cache not set")
	}
	return &API{
		cfg:   cfg,
		repo:  repo,
		cache: cache,
	}, nil
}

func (a *API) HealthCheck(ctx context.Context) error {
	return a.repo.Ping(ctx)
}

func (a *API) SetCacheEnabled(v bool) {
	a.cfg.CacheEnabled = v
}
