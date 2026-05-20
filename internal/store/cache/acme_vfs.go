package cache

import (
	"context"
	"database/sql"
	"errors"
	"path"

	"github.com/rcarmo/gi/internal/store"
	"golang.org/x/crypto/acme/autocert"
)

type VFSCache struct {
	store *store.Store
}

func NewVFSCache(s *store.Store) VFSCache {
	return VFSCache{store: s}
}

func (c VFSCache) Get(ctx context.Context, key string) ([]byte, error) {
	_, value, err := c.store.GetVFSFileContent(ctx, acmeVFSNamespace, acmeVFSPath(key))
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

func (c VFSCache) Put(ctx context.Context, key string, data []byte) error {
	_, err := c.store.SaveVFSFile(ctx, acmeVFSNamespace, acmeVFSPath(key), "application/octet-stream", data, map[string]any{"kind": "acme_autocert"})
	return err
}

func (c VFSCache) Delete(ctx context.Context, key string) error {
	return c.store.DeleteVFSFile(ctx, acmeVFSNamespace, acmeVFSPath(key))
}

func acmeVFSPath(key string) string {
	return path.Join("cache", key)
}

var _ autocert.Cache = VFSCache{}
