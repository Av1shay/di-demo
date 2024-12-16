package uam

import (
	"context"
	"fmt"
	"github.com/Av1shay/di-demo/pkg/types"
	"log"
)

func (a *API) GetItem(ctx context.Context, name string) (types.Item, error) {
	if a.cfg.CacheEnabled {
		var item types.Item
		if err := a.cache.Get(ctx, fmt.Sprintf("item:%s", name), &item); err == nil {
			return item, nil
		}
	}
	item, err := a.itemRepo.GetItemByName(ctx, name)
	if err != nil {
		return types.Item{}, err
	}
	if a.cfg.CacheEnabled {
		if err := a.cache.Set(ctx, cacheKey(name), item, cacheTTL); err != nil {
			log.Println("Failed to save item to cache", err)
		}
	}
	return item, nil
}

func (a *API) CreateItem(ctx context.Context, input types.ItemCreateInput) (types.Item, error) {
	item, err := a.itemRepo.SaveItem(ctx, input)
	if err != nil {
		return types.Item{}, err
	}
	if a.cfg.CacheEnabled {
		if err := a.cache.Set(ctx, cacheKey(input.Name), item, cacheTTL); err != nil {
			log.Println("Failed to save item to cache", err)
		}
	}
	return item, nil
}

func cacheKey(name string) string {
	return fmt.Sprintf("item:%s", name)
}
