package store

import (
	"context"
	"database/sql"

	"golang.org/x/crypto/acme/autocert"
)

const acmeCacheNamespace = "acme/autocert"

type ACMESQLiteCache struct {
	store *Store
}

func NewACMESQLiteCache(s *Store) ACMESQLiteCache {
	return ACMESQLiteCache{store: s}
}

func (c ACMESQLiteCache) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := c.store.GetKV(ctx, acmeCacheNamespace, key)
	if err == sql.ErrNoRows {
		return nil, autocert.ErrCacheMiss
	}
	return value, err
}

func (c ACMESQLiteCache) Put(ctx context.Context, key string, data []byte) error {
	return c.store.PutKV(ctx, acmeCacheNamespace, key, data)
}

func (c ACMESQLiteCache) Delete(ctx context.Context, key string) error {
	return c.store.DeleteKV(ctx, acmeCacheNamespace, key)
}
