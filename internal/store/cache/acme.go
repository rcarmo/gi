package cache

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rcarmo/gi/internal/store"
	"golang.org/x/crypto/acme/autocert"
)

const acmeCacheNamespace = "acme/autocert"

type SQLiteCache struct {
	store *store.Store
}

func NewSQLiteCache(s *store.Store) SQLiteCache {
	return SQLiteCache{store: s}
}

func (c SQLiteCache) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := c.store.GetKV(ctx, acmeCacheNamespace, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, autocert.ErrCacheMiss
		}
		return nil, err
	}
	if len(value) == 0 {
		return nil, autocert.ErrCacheMiss
	}
	return value, nil
}

func (c SQLiteCache) Put(ctx context.Context, key string, data []byte) error {
	return c.store.PutKV(ctx, acmeCacheNamespace, key, data)
}

func (c SQLiteCache) Delete(ctx context.Context, key string) error {
	return c.store.DeleteKV(ctx, acmeCacheNamespace, key)
}

var _ autocert.Cache = SQLiteCache{}
var _ interface{ Delete(context.Context, string) error } = SQLiteCache{}
