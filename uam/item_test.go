package uam

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Av1shay/di-demo/cache/memory"
	"github.com/Av1shay/di-demo/cache/redis"
	"github.com/Av1shay/di-demo/pkg/errs"
	"github.com/Av1shay/di-demo/pkg/test"
	"github.com/Av1shay/di-demo/pkg/types"
	"github.com/Av1shay/di-demo/repositories/mock"
	"github.com/Av1shay/di-demo/repositories/mongo"
	"github.com/Av1shay/di-demo/repositories/mysql"
	"github.com/brianvoe/gofakeit/v7"
	redisdk "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"testing"
	"time"
)

func init() {
	h := slog.NewJSONHandler(io.Discard, nil)
	slog.SetDefault(slog.New(h))
}

const mysqlPassword = "secret"

func TestAPI_GetItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("with_cache", func(t *testing.T) {
		t.Parallel()

		expectedItem := types.Item{
			ID:        gofakeit.UUID(),
			AccountID: gofakeit.UUID(),
			Name:      "test-item-" + gofakeit.LetterN(6),
			Value:     "item-value-" + gofakeit.UUID(),
			CreatedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		}
		mockRepo := mock.Repository{
			SaveItemRes:      expectedItem,
			GetItemByNameRes: expectedItem,
		}
		memoryCache := memory.NewCache()

		api, err := NewAPI(Config{CacheEnabled: true}, &mockRepo, memoryCache)
		require.NoError(t, err)

		createdItem, err := api.CreateItem(ctx, types.ItemCreateInput{
			Name:  expectedItem.Name,
			Value: expectedItem.Value,
		}, expectedItem.AccountID)
		require.NoError(t, err)

		// make sure item is in cache
		var gotCachedItem types.Item
		err = memoryCache.Get(ctx, genItemCacheKey(createdItem.Name, expectedItem.AccountID), &gotCachedItem)
		require.NoError(t, err)
		require.Equal(t, createdItem, gotCachedItem)

		gotItem, err := api.GetItemByName(ctx, createdItem.Name, createdItem.AccountID)
		require.NoError(t, err)

		require.Equal(t, createdItem, gotItem)
	})
}

func TestAPI_GetItem_Integration_MySQL(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(func() { cancel() })

	mysqlAddr, cleanupDB := test.CreateMySQLContainer(ctx, t, mysqlPassword)
	t.Cleanup(func() { require.NoError(t, cleanupDB()) })

	mysqlConn := fmt.Sprintf("root:%s@(%s)/mysql?parseTime=true", mysqlPassword, mysqlAddr)
	db, err := sql.Open("mysql", mysqlConn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, db.Ping())

	_, err = db.Exec(mysql.CreateItemsTableQuery)
	require.NoError(t, err)

	redisAddr, cleanupRedis := test.CreateRedisContainer(ctx, t)
	t.Cleanup(func() { require.NoError(t, cleanupRedis()) })

	t.Run("with_cache", func(t *testing.T) {
		t.Parallel()

		mysqlRepo, err := mysql.NewRepository(mysqlConn)

		rdb := redisdk.NewClient(&redisdk.Options{Addr: redisAddr, DB: 0})
		require.NoError(t, rdb.Ping(ctx).Err())
		cache := redis.NewCache(rdb)

		api, err := NewAPI(Config{CacheEnabled: true}, mysqlRepo, cache)
		require.NoError(t, err)

		accountID := gofakeit.UUID()
		createdItem, err := api.CreateItem(ctx, types.ItemCreateInput{
			Name:  "test-item-" + gofakeit.LetterN(6),
			Value: "item-value-" + gofakeit.UUID(),
		}, accountID)
		require.NoError(t, err)

		// make sure item is in cache
		cachedItem, err := rdb.Get(ctx, redis.BuildKey(genItemCacheKey(createdItem.Name, accountID))).Bytes()
		require.NoError(t, err)
		var gotCachedItem types.Item
		require.NoError(t, json.Unmarshal(cachedItem, &gotCachedItem))
		require.Equal(t, createdItem, gotCachedItem)

		gotItem, err := api.GetItemByName(ctx, createdItem.Name, accountID)
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

		accountID := gofakeit.UUID()
		createdItem, err := api.CreateItem(ctx, types.ItemCreateInput{
			Name:  "test-item-" + gofakeit.LetterN(6),
			Value: "item-value-" + gofakeit.UUID(),
		}, accountID)
		require.NoError(t, err)

		err = rdb.Get(ctx, redis.BuildKey(genItemCacheKey(createdItem.Name, accountID))).Err()
		require.Error(t, err)

		gotItem, err := api.GetItemByName(ctx, createdItem.Name, accountID)
		require.NoError(t, err)

		require.Equal(t, createdItem, gotItem)
	})
}

func TestAPI_ListItems_Integration_MySQL(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(func() { cancel() })

	mysqlAddr, cleanup := test.CreateMySQLContainer(ctx, t, mysqlPassword)
	t.Cleanup(func() { require.NoError(t, cleanup()) })

	mysqlConn := fmt.Sprintf("root:%s@(%s)/mysql?parseTime=true", mysqlPassword, mysqlAddr)
	db, err := sql.Open("mysql", mysqlConn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, db.Ping())

	_, err = db.Exec(mysql.CreateItemsTableQuery)
	require.NoError(t, err)

	mysqlRepo, err := mysql.NewRepository(mysqlConn)

	api, err := NewAPI(Config{CacheEnabled: false}, mysqlRepo, memory.NewCache())
	require.NoError(t, err)

	accountID := gofakeit.UUID()

	gotItems, err := api.ListItems(ctx, types.ListItemsInput{}, accountID)
	require.NoError(t, err)
	require.Empty(t, gotItems)

	const itemsCount = 3
	items := make([]types.Item, itemsCount)
	for i := 0; i < len(items); i++ {
		item, err := api.CreateItem(ctx, types.ItemCreateInput{
			Name:  "test-item-" + gofakeit.LetterN(6),
			Value: "item-value-" + gofakeit.UUID(),
		}, accountID)
		require.NoError(t, err)
		items[i] = item
		time.Sleep(time.Second)
	}

	gotItems, err = api.ListItems(ctx, types.ListItemsInput{}, accountID)
	require.NoError(t, err)
	require.Len(t, gotItems, itemsCount)
	require.ElementsMatch(t, items, gotItems)

	gotItems, err = api.ListItems(ctx, types.ListItemsInput{
		OrderBy: types.OrderByCreatedAt,
		Sort:    types.DESC,
	}, accountID)
	require.NoError(t, err)
	require.Len(t, gotItems, itemsCount)
	require.Equal(t, items[0], gotItems[2])
	require.Equal(t, items[1], gotItems[1])
	require.Equal(t, items[2], gotItems[0])

	_, err = api.UpdateItem(ctx, types.UpdateItemInput{
		ID:    items[0].ID,
		Name:  items[0].Name,
		Value: "item-value-" + gofakeit.UUID(),
	}, accountID)
	require.NoError(t, err)
	gotItems, err = api.ListItems(ctx, types.ListItemsInput{
		OrderBy: types.OrderByUpdatedAt,
		Sort:    types.DESC,
	}, accountID)
	require.NoError(t, err)
	require.Len(t, gotItems, itemsCount)
	require.Equal(t, items[0].ID, gotItems[0].ID)
	require.Equal(t, items[0].Name, gotItems[0].Name)

	gotItems, err = api.ListItems(ctx, types.ListItemsInput{Limit: 2}, accountID)
	require.NoError(t, err)
	require.Len(t, gotItems, 2)

	// check cache
	api.SetCacheEnabled(true)
	gotItems, err = api.ListItems(ctx, types.ListItemsInput{Limit: 2}, accountID)
	require.NoError(t, err)
	require.Len(t, gotItems, 2)

	require.NoError(t, mysqlRepo.Close())

	gotItems, err = api.ListItems(ctx, types.ListItemsInput{Limit: 2}, accountID)
	require.NoError(t, err)
	require.Len(t, gotItems, 2)
}

func TestAPI_ListItems_Integration_Mongo(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(func() { cancel() })

	mongoAddr, cleanup := test.CreateMongoDBContainer(ctx, t)
	t.Cleanup(func() { require.NoError(t, cleanup()) })

	mongoURI := "mongodb://" + mongoAddr
	mongoRepo, err := mongo.NewRepository(mongoURI, "db")
	require.NoError(t, err)

	api, err := NewAPI(Config{CacheEnabled: false}, mongoRepo, memory.NewCache())
	require.NoError(t, err)

	accountID := gofakeit.UUID()

	gotItems, err := api.ListItems(ctx, types.ListItemsInput{}, accountID)
	require.NoError(t, err)
	require.Empty(t, gotItems)

	const itemsCount = 3
	items := make([]types.Item, itemsCount)
	for i := 0; i < len(items); i++ {
		item, err := api.CreateItem(ctx, types.ItemCreateInput{
			Name:  "test-item-" + gofakeit.LetterN(6),
			Value: "item-value-" + gofakeit.UUID(),
		}, accountID)
		require.NoError(t, err)
		items[i] = item
	}

	gotItems, err = api.ListItems(ctx, types.ListItemsInput{}, accountID)
	require.NoError(t, err)
	require.Len(t, gotItems, itemsCount)
	require.ElementsMatch(t, items, gotItems)

	gotItems, err = api.ListItems(ctx, types.ListItemsInput{
		OrderBy: types.OrderByCreatedAt,
		Sort:    types.DESC,
	}, accountID)
	require.NoError(t, err)
	require.Len(t, gotItems, itemsCount)
	require.Equal(t, items[0], gotItems[2])
	require.Equal(t, items[1], gotItems[1])
	require.Equal(t, items[2], gotItems[0])

	_, err = api.UpdateItem(ctx, types.UpdateItemInput{
		ID:    items[0].ID,
		Name:  items[0].Name,
		Value: "item-value-" + gofakeit.UUID(),
	}, accountID)
	require.NoError(t, err)
	gotItems, err = api.ListItems(ctx, types.ListItemsInput{
		OrderBy: types.OrderByUpdatedAt,
		Sort:    types.DESC,
	}, accountID)
	require.NoError(t, err)
	require.Len(t, gotItems, itemsCount)
	require.Equal(t, items[0].ID, gotItems[0].ID)
	require.Equal(t, items[0].Name, gotItems[0].Name)

	gotItems, err = api.ListItems(ctx, types.ListItemsInput{Limit: 2}, accountID)
	require.NoError(t, err)
	require.Len(t, gotItems, 2)

	itemToDelete := items[0]
	err = api.DeleteItem(ctx, itemToDelete.ID, accountID)
	require.NoError(t, err)
	_, err = api.GetItemByName(ctx, itemToDelete.Name, accountID)
	var appErr *errs.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, errs.ErrorCodeNotFound, appErr.Code)
}
