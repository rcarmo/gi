package store

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func openMigrationDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_txlock=immediate")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
func applyMigration(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = Migrate(tx); err != nil {
		return err
	}
	return tx.Commit()
}
func TestWorkspaceMigrationFreshConcurrentAndReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "migration.db")
	a := openMigrationDB(t, path)
	b := openMigrationDB(t, path)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, db := range []*sql.DB{a, b} {
		wg.Add(1)
		go func() { defer wg.Done(); results <- applyMigration(db) }()
	}
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int
	var checksum, appliedAt string
	if err := a.QueryRow("SELECT (SELECT count(*) FROM workspace_index_migrations),checksum,applied_at FROM workspace_index_migrations WHERE version=1").Scan(&count, &checksum, &appliedAt); err != nil || count != len(workspaceMigrations) || len(checksum) != 64 || appliedAt == "" {
		t.Fatal(count, checksum, appliedAt, err)
	}
	if _, err := a.Exec("INSERT INTO workspace_index_workspaces VALUES(1,'retained')"); err != nil {
		t.Fatal(err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	c := openMigrationDB(t, path)
	if err := applyMigration(c); err != nil {
		t.Fatal(err)
	}
	var after string
	if err := c.QueryRow("SELECT applied_at FROM workspace_index_migrations ORDER BY version LIMIT 1").Scan(&after); err != nil || after != appliedAt {
		t.Fatal(after, err)
	}
	if err := c.QueryRow("SELECT count(*) FROM workspace_index_workspaces WHERE root_identity='retained'").Scan(&count); err != nil || count != 1 {
		t.Fatal(count, err)
	}
}

func TestWorkspaceMigrationRollsBackLateFailure(t *testing.T) {
	db := openMigrationDB(t, filepath.Join(t.TempDir(), "rollback.db"))
	// Collide with the last trigger, after tables/FTS/earlier triggers were made.
	if _, err := db.Exec("CREATE TABLE retained(value TEXT); INSERT INTO retained VALUES('keep'); CREATE TRIGGER workspace_index_documents_after_rename AFTER INSERT ON retained BEGIN SELECT 1; END"); err != nil {
		t.Fatal(err)
	}
	if err := applyMigration(db); err == nil {
		t.Fatal("expected trigger collision")
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM sqlite_schema WHERE name LIKE 'workspace_index_%'").Scan(&count); err != nil || count != 1 {
		t.Fatal("partial DDL survived", count, err)
	}
	if _, err := db.Exec("DROP TRIGGER workspace_index_documents_after_rename"); err != nil {
		t.Fatal(err)
	}
	if err := applyMigration(db); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT count(*) FROM retained WHERE value='keep'").Scan(&count); err != nil || count != 1 {
		t.Fatal(count, err)
	}
}

func TestWorkspaceMigrationRejectsUnknownOrDamagedSchemas(t *testing.T) {
	cases := []struct{ name, damage, want string }{
		{"future", "UPDATE workspace_index_migrations SET version=99 WHERE version=2", "unsupported"},
		{"ledger_gap", "DELETE FROM workspace_index_migrations WHERE version=1", "gap"},
		{"invalidation_missing", "DROP TABLE workspace_index_invalidations", "validate"},
		{"checksum", "UPDATE workspace_index_migrations SET checksum='changed'", "checksum"},
		{"trigger_missing", "DROP TRIGGER workspace_index_chunks_insert", "missing"},
		{"trigger_replaced", "DROP TRIGGER workspace_index_chunks_insert; CREATE TRIGGER workspace_index_chunks_insert AFTER INSERT ON workspace_index_chunks BEGIN SELECT 1; END", "definition mismatch"},
		{"table_missing", "DROP TABLE workspace_index_leases", "validate"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := openMigrationDB(t, filepath.Join(t.TempDir(), "damaged.db"))
			if err := applyMigration(db); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(tc.damage); err != nil {
				t.Fatal(err)
			}
			if err := applyMigration(db); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatal(err)
			}
		})
	}
	t.Run("unversioned_candidate", func(t *testing.T) {
		db := openMigrationDB(t, filepath.Join(t.TempDir(), "candidate.db"))
		if _, err := db.Exec(scopedWorkspaceSchemaV1); err != nil {
			t.Fatal(err)
		}
		if err := applyMigration(db); err == nil {
			t.Fatal("unversioned candidate silently adopted")
		}
		var count int
		if err := db.QueryRow("SELECT count(*) FROM sqlite_schema WHERE name='workspace_index_migrations'").Scan(&count); err != nil || count != 0 {
			t.Fatal("ledger leaked", count, err)
		}
	})
}

func seedV1Index(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec(`CREATE TABLE workspace_index_migrations(version INTEGER PRIMARY KEY CHECK(version>0),checksum TEXT NOT NULL,applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')));
 ` + scopedWorkspaceSchemaV1); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO workspace_index_migrations VALUES(1,?,'kept-date')", fmt.Sprintf("%x", sha256.Sum256([]byte(scopedWorkspaceSchemaV1)))); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO workspace_index_workspaces VALUES(1,'workspace');
 INSERT INTO workspace_index_scopes VALUES(1,'notes','fingerprint','["notes"]','ready',4,1,100,100,'');
 INSERT INTO workspace_index_documents VALUES(1,1,'notes/kept.md','text','',5,100,'hash','v1',100);
 INSERT INTO workspace_index_memberships VALUES(1,'notes',1,4);
 INSERT INTO workspace_index_chunks VALUES(1,1,0,0,5,1,1,'','kept flowers');`); err != nil {
		t.Fatal(err)
	}
}
func TestWorkspaceV2MigrationBackfillReopenAndRollback(t *testing.T) {
	db := openMigrationDB(t, filepath.Join(t.TempDir(), "upgrade.db"))
	seedV1Index(t, db)
	// A ledger write rejection happens after v2 DDL/backfill. All must roll back.
	if _, err := db.Exec(`CREATE TRIGGER reject_v2 BEFORE INSERT ON workspace_index_migrations WHEN new.version=2 BEGIN SELECT raise(abort,'v2 rejected'); END`); err != nil {
		t.Fatal(err)
	}
	if err := applyMigration(db); err == nil {
		t.Fatal("expected late rejection")
	}
	var n int
	if err := db.QueryRow("SELECT count(*) FROM sqlite_schema WHERE name='workspace_index_invalidations'").Scan(&n); err != nil || n != 0 {
		t.Fatal("DDL leaked", n, err)
	}
	if _, err := db.Exec("DROP TRIGGER reject_v2"); err != nil {
		t.Fatal(err)
	}
	for pass := 0; pass < 2; pass++ {
		if err := applyMigration(db); err != nil {
			t.Fatal(err)
		}
		var requested, ack int64
		if err := db.QueryRow("SELECT requested_revision,acknowledged_revision FROM workspace_index_invalidations WHERE workspace_id=1 AND scope='notes'").Scan(&requested, &ack); err != nil || requested != 0 || ack != 0 {
			t.Fatal(requested, ack, err)
		}
		var state, applied string
		var generation int
		if err := db.QueryRow("SELECT state,committed_generation FROM workspace_index_scopes").Scan(&state, &generation); err != nil || state != "ready" || generation != 4 {
			t.Fatal(state, generation, err)
		}
		if err := db.QueryRow("SELECT applied_at FROM workspace_index_migrations WHERE version=1").Scan(&applied); err != nil || applied != "kept-date" {
			t.Fatal(applied, err)
		}
		if err := db.QueryRow("SELECT count(*) FROM workspace_index_fts WHERE workspace_index_fts MATCH 'flowers'").Scan(&n); err != nil || n != 1 {
			t.Fatal(n, err)
		}
		if _, err := db.Exec("INSERT INTO workspace_index_fts(workspace_index_fts,rank) VALUES('integrity-check',1)"); err != nil {
			t.Fatal(err)
		}
	}
}
