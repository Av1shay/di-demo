package uam

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/Av1shay/di-demo/pkg/log"
	"github.com/Av1shay/di-demo/pkg/types"
)

func (a *API) GetItemByName(ctx context.Context, name, accountID string) (types.Item, error) {
	if a.cfg.CacheEnabled {
		var item types.Item
		if err := a.cache.Get(ctx, genItemCacheKey(name, accountID), &item); err == nil {
			return item, nil
		}
	}
	item, err := a.repo.GetItemByName(ctx, name, accountID)
	if err != nil {
		return types.Item{}, err
	}
	if a.cfg.CacheEnabled {
		if err := a.cache.Set(ctx, genItemCacheKey(name, accountID), item, itemCacheTTL); err != nil {
			log.Errorf(ctx, "Failed to save item to cache: %v", err)
		}
	}
	return item, nil
}

func (a *API) CreateItem(ctx context.Context, input types.ItemCreateInput, accountID string) (types.Item, error) {
	item, err := a.repo.SaveItem(ctx, input, accountID)
	if err != nil {
		return types.Item{}, err
	}
	if a.cfg.CacheEnabled {
		if err := a.cache.Set(ctx, genItemCacheKey(input.Name, accountID), item, itemCacheTTL); err != nil {
			log.Errorf(ctx, "Failed to save item to cache: %v", err)
		}
	}
	return item, nil
}

func (a *API) ListItems(ctx context.Context, input types.ListItemsInput, accountID string) ([]types.Item, error) {
	cacheKey := ""
	if a.cfg.CacheEnabled {
		if k, err := genListCacheKey(input, accountID); err == nil {
			cacheKey = k
			var items []types.Item
			if err := a.cache.Get(ctx, cacheKey, &items); err == nil {
				return items, nil
			}
		}
	}
	items, err := a.repo.ListItems(ctx, input, accountID)
	if err != nil {
		return nil, err
	}
	if a.cfg.CacheEnabled && cacheKey != "" {
		if err := a.cache.Set(ctx, cacheKey, items, listCacheTTL); err != nil {
			log.Errorf(ctx, "Failed to save list items to cache: %v", err)
		}
	}
	return items, nil
}

func (a *API) UpdateItem(ctx context.Context, input types.UpdateItemInput, accountID string) (types.Item, error) {
	item, err := a.repo.UpdateItem(ctx, input, accountID)
	if err != nil {
		return types.Item{}, err
	}
	if a.cfg.CacheEnabled {
		if input.Name != item.Name {
			_ = a.cache.Delete(ctx, genItemCacheKey(input.Name, accountID))
		}
		if err := a.cache.Set(ctx, genItemCacheKey(item.Name, accountID), item, itemCacheTTL); err != nil {
			log.Errorf(ctx, "Failed to save item to cache: %v", err)
		}
	}
	return item, nil
}

func (a *API) DeleteItem(ctx context.Context, id, accountID string) error {
	return a.repo.DeleteItem(ctx, id, accountID)
}

func genItemCacheKey(name, accountID string) string {
	return fmt.Sprintf("item:%s:%s", name, accountID)
}

func genListCacheKey(input types.ListItemsInput, accountID string) (string, error) {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to serialize input: %w", err)
	}
	hash := sha256.Sum256(append(inputBytes, []byte(accountID)...))
	return fmt.Sprintf("%x", hash), nil
}
