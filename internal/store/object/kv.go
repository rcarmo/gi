package object

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func PutKV(ctx context.Context, db *sql.DB, namespace, key string, value []byte) error {
	namespace = strings.TrimSpace(namespace)
	key = strings.TrimSpace(key)
	if namespace == "" {
		return fmt.Errorf("put kv: missing namespace")
	}
	if key == "" {
		return fmt.Errorf("put kv: missing key")
	}
	if value == nil {
		value = []byte{}
	}
	_, err := db.ExecContext(ctx, `
		insert into kv_store (namespace, key, value, created_at, updated_at)
		values (?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now'), strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		on conflict(namespace, key) do update set
			value = excluded.value,
			updated_at = excluded.updated_at
	`, namespace, key, value)
	if err != nil {
		return fmt.Errorf("put kv: %w", err)
	}
	return nil
}

func GetKV(ctx context.Context, db *sql.DB, namespace, key string) ([]byte, error) {
	namespace = strings.TrimSpace(namespace)
	key = strings.TrimSpace(key)
	if namespace == "" {
		return nil, fmt.Errorf("get kv: missing namespace")
	}
	if key == "" {
		return nil, fmt.Errorf("get kv: missing key")
	}
	var value []byte
	if err := db.QueryRowContext(ctx, `select value from kv_store where namespace = ? and key = ?`, namespace, key).Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get kv: %w", err)
	}
	return append([]byte(nil), value...), nil
}

func DeleteKV(ctx context.Context, db *sql.DB, namespace, key string) error {
	namespace = strings.TrimSpace(namespace)
	key = strings.TrimSpace(key)
	if namespace == "" {
		return fmt.Errorf("delete kv: missing namespace")
	}
	if key == "" {
		return fmt.Errorf("delete kv: missing key")
	}
	if _, err := db.ExecContext(ctx, `delete from kv_store where namespace = ? and key = ?`, namespace, key); err != nil {
		return fmt.Errorf("delete kv: %w", err)
	}
	return nil
}
