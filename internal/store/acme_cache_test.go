package store_test

import (
	"context"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	storecache "github.com/rcarmo/gi/internal/store/cache"
	"golang.org/x/crypto/acme/autocert"
)

func TestACMEKVCache(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	cache := storecache.NewSQLiteCache(s)
	ctx := context.Background()
	if _, err := cache.Get(ctx, "missing"); err != autocert.ErrCacheMiss {
		t.Fatalf("missing error = %v, want ErrCacheMiss", err)
	}
	if err := cache.Put(ctx, "cert-key", []byte("cert-data")); err != nil {
		t.Fatalf("put: %v", err)
	}
	got, err := cache.Get(ctx, "cert-key")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(got) != "cert-data" {
		t.Fatalf("got %q", got)
	}
	if err := cache.Delete(ctx, "cert-key"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := cache.Get(ctx, "cert-key"); err != autocert.ErrCacheMiss {
		t.Fatalf("deleted error = %v, want ErrCacheMiss", err)
	}
}

func TestACMEVFSCache(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	cache := storecache.NewVFSCache(s)
	ctx := context.Background()
	if _, err := cache.Get(ctx, "missing"); err != autocert.ErrCacheMiss {
		t.Fatalf("missing error = %v, want ErrCacheMiss", err)
	}
	if err := cache.Put(ctx, "domain/cert", []byte("cert-data")); err != nil {
		t.Fatalf("put: %v", err)
	}
	got, err := cache.Get(ctx, "domain/cert")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(got) != "cert-data" {
		t.Fatalf("got %q", got)
	}
	if err := cache.Delete(ctx, "domain/cert"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := cache.Get(ctx, "domain/cert"); err != autocert.ErrCacheMiss {
		t.Fatalf("deleted error = %v, want ErrCacheMiss", err)
	}
}
