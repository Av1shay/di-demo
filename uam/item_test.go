package uam

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Av1shay/di-demo/cache/memory"
	redis "github.com/Av1shay/di-demo/cache/redis"
	"github.com/Av1shay/di-demo/pkg/test"
	"github.com/Av1shay/di-demo/pkg/types"
	"github.com/Av1shay/di-demo/repositories/mysql"
	"github.com/brianvoe/gofakeit/v7"
	redisdk "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const mysqlPassword = "secret"

func TestAPI_GetItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(func() { cancel() })

	mysqlAddr := test.CreateMySQLContainer(t, mysqlPassword)
	mysqlConn := fmt.Sprintf("root:%s@(%s)/mysql?parseTime=true", mysqlPassword, mysqlAddr)
	db, err := sql.Open("mysql", mysqlConn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, db.Ping())

	test.CreateMySQLItemsTable(t, db)

	redisAddr := test.CreateRedisContainer(t)

	t.Run("with_cache", func(t *testing.T) {
		t.Parallel()

		mysqlRepo, err := mysql.NewRepository(mysqlConn)

		rdb := redisdk.NewClient(&redisdk.Options{Addr: redisAddr, DB: 0})
		require.NoError(t, rdb.Ping(ctx).Err())
		cache := redis.NewCache(rdb)

		api, err := NewAPI(Config{CacheEnabled: true}, mysqlRepo, cache)
		require.NoError(t, err)

		createdItem, err := api.CreateItem(ctx, types.ItemCreateInput{
			Name:  "test-item-" + gofakeit.LetterN(6),
			Value: "item-value-" + gofakeit.UUID(),
		})
		require.NoError(t, err)

		// make sure item is in cache
		cachedItem, err := rdb.Get(ctx, redis.BuildKey(cacheKey(createdItem.Name))).Bytes()
		require.NoError(t, err)
		var gotCachedItem types.Item
		require.NoError(t, json.Unmarshal(cachedItem, &gotCachedItem))
		require.Equal(t, createdItem, gotCachedItem)

		gotItem, err := api.GetItem(ctx, createdItem.Name)
		require.NoError(t, err)

		require.Equal(t, createdItem, gotItem)
	})

	t.Run("without_cache", func(t *testing.T) {
		t.Parallel()

		mysqlRepo, err := mysql.NewRepository(mysqlConn)

		rdb := redisdk.NewClient(&redisdk.Options{Addr: redisAddr, DB: 0})
		require.NoError(t, rdb.Ping(ctx).Err())
		cache := redis.NewCache(rdb)

		api, err := NewAPI(Config{CacheEnabled: false}, mysqlRepo, cache)
		require.NoError(t, err)

		createdItem, err := api.CreateItem(ctx, types.ItemCreateInput{
			Name:  "test-item-" + gofakeit.LetterN(6),
			Value: "item-value-" + gofakeit.UUID(),
		})
		require.NoError(t, err)

		err = rdb.Get(ctx, redis.BuildKey(cacheKey(createdItem.Name))).Err()
		require.Error(t, err)

		gotItem, err := api.GetItem(ctx, createdItem.Name)
		require.NoError(t, err)

		require.Equal(t, createdItem, gotItem)
	})
}

func TestAPI_GetItemV2(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(func() { cancel() })

	t.Run("with_cache", func(t *testing.T) {
		t.Parallel()

		expectedItem := types.Item{
			ID:        gofakeit.UUID(),
			Name:      "test-item-" + gofakeit.LetterN(6),
			Value:     "item-value-" + gofakeit.UUID(),
			CreatedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		}
		mockRepo := &mockedRepo{
			getItemOut:  expectedItem,
			saveItemOut: expectedItem,
		}

		memoryCache := memory.NewCache()

		api, err := NewAPI(Config{CacheEnabled: true}, mockRepo, memoryCache)
		require.NoError(t, err)

		createdItem, err := api.CreateItem(ctx, types.ItemCreateInput{
			Name:  expectedItem.Name,
			Value: expectedItem.Value,
		})
		require.NoError(t, err)

		// make sure item is in cache
		var gotCachedItem types.Item
		err = memoryCache.Get(ctx, cacheKey(createdItem.Name), &gotCachedItem)
		require.NoError(t, err)
		require.Equal(t, createdItem, gotCachedItem)

		gotItem, err := api.GetItem(ctx, createdItem.Name)
		require.NoError(t, err)

		require.Equal(t, createdItem, gotItem)
	})
}

type mockedRepo struct {
	getItemOut    types.Item
	saveItemOut   types.Item
	listItemsOut  []types.Item
	deleteItemOut error
}

func (m mockedRepo) GetItemByName(ctx context.Context, name string) (types.Item, error) {
	return m.getItemOut, nil
}

func (m mockedRepo) SaveItem(ctx context.Context, input types.ItemCreateInput) (types.Item, error) {
	return m.getItemOut, nil

}

func (m mockedRepo) ListItems(ctx context.Context) ([]types.Item, error) {
	return m.listItemsOut, nil
}

func (m mockedRepo) DeleteItem(ctx context.Context, id string) error {
	return m.deleteItemOut
}
