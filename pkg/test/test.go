package test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"net"
	"testing"
)

func CreateMySQLContainer(ctx context.Context, t *testing.T, password string) (string, func() error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal(err)
	}

	err = pool.Client.PingWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}

	resource, err := pool.Run("mysql", "8", []string{"MYSQL_ROOT_PASSWORD=" + password})
	if err != nil {
		t.Fatal(err)
	}
	cleanup := func() error {
		return pool.Purge(resource)
	}

	addr := net.JoinHostPort("localhost", resource.GetPort("3306/tcp"))

	if err := pool.Retry(func() error {
		db, err := sql.Open("mysql", fmt.Sprintf("root:%s@(%s)/mysql", password, addr))
		if err != nil {
			return err
		}
		return db.PingContext(ctx)
	}); err != nil {
		_ = cleanup()
		t.Fatal(err)
	}

	return addr, cleanup
}

func CreateRedisContainer(ctx context.Context, t *testing.T) (string, func() error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal(err)
	}

	err = pool.Client.PingWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}

	resource, err := pool.Run("redis", "7.4.1-alpine", nil)
	if err != nil {
		t.Fatal(err)
	}
	cleanup := func() error {
		return pool.Purge(resource)
	}

	addr := net.JoinHostPort("localhost", resource.GetPort("6379/tcp"))

	if err := pool.Retry(func() error {
		client := redis.NewClient(&redis.Options{Addr: addr})
		defer client.Close()
		return client.Ping(ctx).Err()
	}); err != nil {
		_ = cleanup()
		t.Fatal(err)
	}

	return addr, cleanup
}

func CreateMongoDBContainer(ctx context.Context, t *testing.T) (string, func() error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal(err)
	}

	err = pool.Client.PingWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}

	resource, err := pool.Run("mongo", "5.0", nil)
	if err != nil {
		t.Fatal(err)
	}
	cleanup := func() error {
		return pool.Purge(resource)
	}

	addr := net.JoinHostPort("localhost", resource.GetPort("27017/tcp"))

	if err := pool.Retry(func() error {
		client, err := mongo.Connect(options.Client().ApplyURI("mongodb://" + addr))
		if err != nil {
			return err
		}
		return client.Ping(ctx, nil)
	}); err != nil {
		_ = cleanup()
		t.Fatal(err)
	}

	return addr, cleanup
}
