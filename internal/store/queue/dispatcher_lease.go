package queue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	inboundDispatcherLeaseNamespace = "runtime_leases"
	inboundDispatcherLeaseKey       = "inbound_dispatcher"
)

type inboundDispatcherLeaseRecord struct {
	Owner     string `json:"owner"`
	ExpiresAt string `json:"expires_at"`
}

func AcquireInboundDispatcherLease(ctx context.Context, db *sql.DB, owner string, ttl time.Duration) (bool, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return false, fmt.Errorf("acquire inbound dispatcher lease: owner is required")
	}
	if ttl <= 0 {
		return false, fmt.Errorf("acquire inbound dispatcher lease: ttl must be > 0")
	}
	now := time.Now().UTC()
	expiresAt := now.Add(ttl).Format(time.RFC3339Nano)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("acquire inbound dispatcher lease begin tx: %w", err)
	}
	defer tx.Rollback()
	var value []byte
	err = tx.QueryRowContext(ctx, `select value from kv_store where namespace = ? and key = ?`, inboundDispatcherLeaseNamespace, inboundDispatcherLeaseKey).Scan(&value)
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("acquire inbound dispatcher lease load: %w", err)
	}
	if err == nil {
		var current inboundDispatcherLeaseRecord
		if unmarshalErr := json.Unmarshal(value, &current); unmarshalErr != nil {
			return false, fmt.Errorf("acquire inbound dispatcher lease decode: %w", unmarshalErr)
		}
		currentOwner := strings.TrimSpace(current.Owner)
		currentExpiry, parseErr := time.Parse(time.RFC3339Nano, strings.TrimSpace(current.ExpiresAt))
		if parseErr != nil {
			return false, fmt.Errorf("acquire inbound dispatcher lease parse expiry: %w", parseErr)
		}
		if currentOwner != "" && currentOwner != owner && currentExpiry.After(now) {
			return false, nil
		}
	}
	blob, err := json.Marshal(inboundDispatcherLeaseRecord{Owner: owner, ExpiresAt: expiresAt})
	if err != nil {
		return false, fmt.Errorf("acquire inbound dispatcher lease encode: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		insert into kv_store (namespace, key, value, created_at, updated_at)
		values (?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now'), strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		on conflict(namespace, key) do update set
			value = excluded.value,
			updated_at = excluded.updated_at
	`, inboundDispatcherLeaseNamespace, inboundDispatcherLeaseKey, blob); err != nil {
		return false, fmt.Errorf("acquire inbound dispatcher lease store: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("acquire inbound dispatcher lease commit: %w", err)
	}
	return true, nil
}

func ReleaseInboundDispatcherLease(ctx context.Context, db *sql.DB, owner string) error {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return fmt.Errorf("release inbound dispatcher lease: owner is required")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("release inbound dispatcher lease begin tx: %w", err)
	}
	defer tx.Rollback()
	var value []byte
	err = tx.QueryRowContext(ctx, `select value from kv_store where namespace = ? and key = ?`, inboundDispatcherLeaseNamespace, inboundDispatcherLeaseKey).Scan(&value)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("release inbound dispatcher lease load: %w", err)
	}
	var current inboundDispatcherLeaseRecord
	if err := json.Unmarshal(value, &current); err != nil {
		return fmt.Errorf("release inbound dispatcher lease decode: %w", err)
	}
	if strings.TrimSpace(current.Owner) != owner {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `delete from kv_store where namespace = ? and key = ?`, inboundDispatcherLeaseNamespace, inboundDispatcherLeaseKey); err != nil {
		return fmt.Errorf("release inbound dispatcher lease delete: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("release inbound dispatcher lease commit: %w", err)
	}
	return nil
}
