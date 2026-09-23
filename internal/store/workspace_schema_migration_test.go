package store

import (
	"database/sql"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	searchstore "github.com/rcarmo/gi/internal/search/store"
)

// Compare every column/value, without printing stored message/media contents.
func databaseRows(t *testing.T, db *sql.DB, table string) [][]any {
	t.Helper()
	rows, err := db.Query("SELECT * FROM " + table + " ORDER BY rowid")
	if err != nil {
		t.Fatal(table, err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	var snapshot [][]any
	for rows.Next() {
		values := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for j := range values {
			ptrs[j] = &values[j]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatal(err)
		}
		for j, v := range values {
			if raw, ok := v.([]byte); ok {
				values[j] = append([]byte(nil), raw...)
			}
		}
		snapshot = append(snapshot, values)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func TestWorkspaceMigrationPreservesLegacySearchAndRuntimeData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "historical.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(legacySchema); err != nil {
		t.Fatal(err)
	}
	// Upgrade the core legacy schema first, then remove the new index namespace
	// to construct a pre-index database with otherwise current runtime tables.
	if err := initSchema(db); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query("SELECT name,type FROM sqlite_schema WHERE name LIKE 'workspace_index_%' AND type IN ('trigger','view')")
	if err != nil {
		t.Fatal(err)
	}
	var drops []string
	for rows.Next() {
		var name, kind string
		if err := rows.Scan(&name, &kind); err != nil {
			t.Fatal(err)
		}
		drops = append(drops, "DROP "+kind+" "+name)
	}
	rows.Close()
	for _, query := range drops {
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"invalidations", "fts", "memberships", "chunks", "documents", "leases", "scopes", "workspaces", "migrations"} {
		if _, err := db.Exec("DROP TABLE workspace_index_" + name); err != nil {
			t.Fatal(err)
		}
	}
	if err := searchstore.InitSchema(db); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO workspace_documents(id,path,kind,size_bytes,mtime_ns,content_hash,chunk_count,indexed_at_ms) VALUES(7,'legacy.txt','text',5,12,'hash',1,33);
 INSERT INTO workspace_chunks(id,document_id,chunk_index,start_byte,end_byte,content,embedding_version) VALUES(8,7,0,0,5,'legacy lexical','old');
 INSERT INTO workspace_chunks_fts(rowid,content,path) VALUES(8,'legacy lexical','legacy.txt');
 INSERT INTO workspace_index_meta VALUES('status','{"state":"ready","indexed_file_count":1}');
 INSERT INTO media(session_id,filename,content_type,metadata_json,original_size,compressed_size,compressed,content,created_at,updated_at) VALUES('A','kept.bin','application/octet-stream','{}',4,4,0,X'0001FEFF','original','original');
 INSERT INTO vfs_files(namespace,path,content_type,metadata_json,original_size,compressed_size,compressed,content,created_at,updated_at) VALUES('workspace','kept','text/plain','{}',4,4,0,X'74657374','original','original');
 INSERT INTO kv_store VALUES('kept','key',X'0100FE','original','original');`); err != nil {
		t.Fatal(err)
	}
	tables := []string{"sessions", "session_identities", "messages", "turns", "turn_events", "media", "vfs_files", "kv_store", "workspace_documents", "workspace_chunks", "workspace_chunks_fts", "workspace_index_meta"}
	before := map[string][][]any{}
	for _, table := range tables {
		before[table] = databaseRows(t, db, table)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	for pass := 0; pass < 2; pass++ {
		s, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, table := range tables {
			if !reflect.DeepEqual(before[table], databaseRows(t, s.DB(), table)) {
				s.Close()
				t.Fatalf("%s changed on migration/reopen %d", table, pass)
			}
		}
		for _, table := range []string{"workspace_index_workspaces", "workspace_index_scopes", "workspace_index_documents", "workspace_index_chunks", "workspace_index_memberships", "workspace_index_leases"} {
			var n int
			if err := s.DB().QueryRow("SELECT count(*) FROM " + table).Scan(&n); err != nil || n != 0 {
				t.Fatal(table, n, err)
			}
		}
		var hits int
		if err := s.DB().QueryRow("SELECT count(*) FROM workspace_chunks_fts WHERE workspace_chunks_fts MATCH 'lexical'").Scan(&hits); err != nil || hits != 1 {
			t.Fatal(hits, err)
		}
		check, err := s.DB().Query("PRAGMA foreign_key_check")
		if err != nil {
			t.Fatal(err)
		}
		if check.Next() {
			t.Fatal("foreign-key violation")
		}
		check.Close()
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestWorkspaceMigrationFailureRollsBackCoreLegacyUpgrade(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rollback.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(legacySchema + `CREATE TRIGGER workspace_index_documents_after_rename AFTER INSERT ON messages BEGIN SELECT 1; END;`); err != nil {
		t.Fatal(err)
	}
	messages := databaseRows(t, db, "messages")
	if err := initSchema(db); err == nil || !strings.Contains(err.Error(), "workspace") {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM pragma_table_info('turns') WHERE name='phase'").Scan(&count); err != nil || count != 0 {
		t.Fatal("core ALTER leaked", count, err)
	}
	if err := db.QueryRow("SELECT count(*) FROM sqlite_schema WHERE name='workspace_index_migrations' OR name='steering_queue'").Scan(&count); err != nil || count != 0 {
		t.Fatal("schema leaked", count, err)
	}
	if !reflect.DeepEqual(messages, databaseRows(t, db, "messages")) {
		t.Fatal("history changed")
	}
	if _, err := db.Exec("DROP TRIGGER workspace_index_documents_after_rename"); err != nil {
		t.Fatal(err)
	}
	if err := initSchema(db); err != nil {
		t.Fatal(err)
	}
}
