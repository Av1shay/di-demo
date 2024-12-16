package test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func CreateMySQLContainer(t *testing.T, password string) string {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal(err)
	}

	err = pool.Client.Ping()
	if err != nil {
		t.Fatal(err)
	}

	resource, err := pool.Run("mysql", "8", []string{"MYSQL_ROOT_PASSWORD=" + password})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := pool.Purge(resource); err != nil {
			t.Fatalf("Could not purge resource: %s", err)
		}
	})

	addr := net.JoinHostPort("localhost", resource.GetPort("3306/tcp"))

	if err := pool.Retry(func() error {
		db, err := sql.Open("mysql", fmt.Sprintf("root:%s@(%s)/mysql", password, addr))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		t.Fatal(err)
	}

	return addr
}

func CreateRedisContainer(t *testing.T) string {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal(err)
	}

	err = pool.Client.Ping()
	if err != nil {
		t.Fatal(err)
	}

	resource, err := pool.Run("redis", "7.4.1-alpine", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := pool.Purge(resource); err != nil {
			t.Fatalf("Could not purge resource: %s", err)
		}
	})

	addr := net.JoinHostPort("localhost", resource.GetPort("6379/tcp"))

	if err := pool.Retry(func() error {
		client := redis.NewClient(&redis.Options{Addr: addr})
		defer client.Close()
		return client.Ping(context.Background()).Err()
	}); err != nil {
		t.Fatal(err)
	}

	return addr
}

func CreateMySQLItemsTable(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS items (
		id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		value VARCHAR(255) NULL,
		version INT NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX (name)
	)`)
	require.NoError(t, err)
}
