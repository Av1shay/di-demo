package memory

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
)

type cacheVal struct {
	ID  int    `json:"id"`
	Foo string `json:"foo"`
}

func TestCache_Get(t *testing.T) {
	ctx := context.Background()
	c := NewCache()
	key1, val1 := "somekey1", cacheVal{123, "baz"}
	key2, val2 := "somekey2", cacheVal{652, "xyz asdasd asd asd ass"}

	if err := c.Set(ctx, key1, val1, time.Second); err != nil {
		t.Fatal(err)
	}
	var cacheRes1 cacheVal
	if err := c.Get(ctx, key1, &cacheRes1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(val1, cacheRes1) {
		t.Fatalf("cached results are not equal, want %+v, got %+v", val1, cacheRes1)
	}
	if err := c.Delete(ctx, key1); err != nil {
		t.Fatal(err)
	}
	err := c.Get(ctx, key1, &cacheRes1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "key not found") {
		t.Fatal("expected error to contain 'key not found'")
	}

	if err := c.Set(ctx, key2, val2, time.Millisecond); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond)
	var cacheRes2 cacheVal
	err = c.Get(ctx, key2, &cacheRes2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "key expired") {
		t.Fatal("expected error to contain 'key expired'")
	}
}
