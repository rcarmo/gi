package store

import (
	"context"
	"database/sql"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

const acmeVFSNamespace = "acme-autocert"

type ACMEVFSCache struct {
	store *Store
}

func NewACMEVFSCache(s *Store) ACMEVFSCache {
	return ACMEVFSCache{store: s}
}

func (c ACMEVFSCache) Get(ctx context.Context, key string) ([]byte, error) {
	_, value, err := c.store.GetVFSFileContent(ctx, acmeVFSNamespace, acmeVFSPath(key))
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "sql: no rows") {
			return nil, autocert.ErrCacheMiss
		}
		return nil, err
	}
	return value, nil
}

func (c ACMEVFSCache) Put(ctx context.Context, key string, data []byte) error {
	_, err := c.store.SaveVFSFile(ctx, acmeVFSNamespace, acmeVFSPath(key), "application/octet-stream", data, map[string]any{"kind": "acme_autocert"})
	return err
}

func (c ACMEVFSCache) Delete(ctx context.Context, key string) error {
	return c.store.DeleteVFSFile(ctx, acmeVFSNamespace, acmeVFSPath(key))
}

func acmeVFSPath(key string) string {
	key = strings.Trim(strings.ReplaceAll(key, "\\", "/"), "/")
	if key == "" {
		return "cache-item"
	}
	return key
}
