package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Av1shay/di-demo/cache/redis"
	"github.com/Av1shay/di-demo/pkg/test"
	"github.com/Av1shay/di-demo/pkg/types"
	repomysql "github.com/Av1shay/di-demo/repositories/mysql"
	"github.com/Av1shay/di-demo/uam"
	"github.com/brianvoe/gofakeit/v7"
	redissdk "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"io"
	"net/http/httptest"
	"testing"
	"time"
)

const mysqlPassword = "secret"

func TestServer_GetItem(t *testing.T) {
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

	repo, err := repomysql.NewRepository(mysqlConn)
	if err != nil {
		t.Fatal("Failed to create user repository client", err)
	}

	rdb := redissdk.NewClient(&redissdk.Options{Addr: redisAddr, DB: 0})
	require.NoError(t, rdb.Ping(ctx).Err())

	cache := redis.NewCache(rdb)

	uamAPI, err := uam.NewAPI(uam.Config{}, repo, cache)
	require.NoError(t, err)

	serv := New(uamAPI)
	serv.MountHandlers()
	ts := httptest.NewServer(serv.Router())
	t.Cleanup(ts.Close)

	t.Run("test_not_exist", func(t *testing.T) {
		t.Parallel()

		resp, err := ts.Client().Get(ts.URL + "/item/-1")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, 404, resp.StatusCode)
	})

	t.Run("test_ok", func(t *testing.T) {
		t.Parallel()

		createdItem, err := uamAPI.CreateItem(ctx, types.ItemCreateInput{
			Name:  "test-item-" + gofakeit.LetterN(6),
			Value: "test-value-" + gofakeit.UUID(),
		})

		resp, err := ts.Client().Get(fmt.Sprintf("%s/item/%s", ts.URL, createdItem.Name))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, 200, resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		var gotItem types.Item
		require.NoError(t, json.Unmarshal(b, &gotItem))
		require.Equal(t, createdItem, gotItem)
	})
}
