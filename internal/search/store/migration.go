package store

import (
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed migrations/001_scoped_workspace_index.sql
var scopedWorkspaceSchemaV1 string

const workspaceSchemaVersion = 1

// Migrate runs inside the caller's existing startup transaction. It never
// commits, creates workspace identities, scans files or imports legacy rows.
func Migrate(tx *sql.Tx) error {
	if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS workspace_index_migrations (
  version INTEGER PRIMARY KEY CHECK(version > 0),
  checksum TEXT NOT NULL,
  applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
 )`); err != nil {
		return fmt.Errorf("workspace migration ledger: %w", err)
	}
	rows, err := tx.Query("SELECT version,checksum FROM workspace_index_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("read workspace migration ledger: %w", err)
	}
	applied := false
	sum := fmt.Sprintf("%x", sha256.Sum256([]byte(scopedWorkspaceSchemaV1)))
	for rows.Next() {
		var version int
		var checksum string
		if err = rows.Scan(&version, &checksum); err != nil {
			rows.Close()
			return err
		}
		if version != workspaceSchemaVersion {
			rows.Close()
			return fmt.Errorf("unsupported workspace index schema version %d", version)
		}
		if checksum != sum {
			rows.Close()
			return fmt.Errorf("workspace index schema v%d checksum mismatch", version)
		}
		applied = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if !applied {
		// CREATE without IF NOT EXISTS detects unversioned partial/candidate tables.
		// They cannot silently be adopted as a complete migrated schema.
		if _, err = tx.Exec(scopedWorkspaceSchemaV1); err != nil {
			return fmt.Errorf("apply workspace index schema v1: %w", err)
		}
		if _, err = tx.Exec("INSERT INTO workspace_index_migrations(version,checksum) VALUES (?,?)", workspaceSchemaVersion, sum); err != nil {
			return err
		}
	}
	return validateWorkspaceSchema(tx)
}

// The ledger authenticates migration history, not arbitrary later DDL. Validate
// required columns and objects on reopen as well; do not silently recreate a
// missing trigger and leave existing FTS rows inconsistent.
func validateWorkspaceSchema(tx *sql.Tx) error {
	projections := map[string]string{
		"workspace_index_workspaces":    "id,root_identity",
		"workspace_index_scopes":        "workspace_id,scope,config_hash,roots_json,state,committed_generation,indexed_file_count,last_indexed_at_ms,updated_at_ms,last_error",
		"workspace_index_leases":        "workspace_id,owner_token,expires_at_ms",
		"workspace_index_documents":     "id,workspace_id,path,kind,language,size_bytes,mtime_ns,content_hash,chunker_version,indexed_at_ms",
		"workspace_index_memberships":   "workspace_id,scope,document_id,seen_generation",
		"workspace_index_chunks":        "id,document_id,chunk_index,start_byte,end_byte,start_line,end_line,heading,content",
		"workspace_index_chunk_content": "id,content,heading,path,language",
		"workspace_index_fts":           "rowid,content,heading,path,language",
	}
	for table, columns := range projections {
		rows, err := tx.Query("SELECT " + columns + " FROM " + table + " LIMIT 0")
		if err != nil {
			return fmt.Errorf("validate %s: %w", table, err)
		}
		rows.Close()
	}
	objects := map[string]string{
		"workspace_index_workspaces": "table", "workspace_index_scopes": "table", "workspace_index_leases": "table",
		"workspace_index_documents": "table", "workspace_index_memberships": "table", "workspace_index_chunks": "table",
		"workspace_index_chunk_content": "view", "workspace_index_fts": "table",
		"workspace_index_chunks_by_document": "index", "workspace_index_memberships_by_document": "index",
		"workspace_index_chunks_insert": "trigger", "workspace_index_chunks_delete": "trigger",
		"workspace_index_chunks_before_update": "trigger", "workspace_index_chunks_after_update": "trigger",
		"workspace_index_documents_delete": "trigger", "workspace_index_documents_before_rename": "trigger", "workspace_index_documents_after_rename": "trigger",
	}
	for name, kind := range objects {
		var actual, ddl string
		if err := tx.QueryRow("SELECT type,sql FROM sqlite_schema WHERE name=?", name).Scan(&actual, &ddl); err != nil {
			return fmt.Errorf("missing workspace schema object %s: %w", name, err)
		}
		if actual != kind {
			return fmt.Errorf("workspace schema object %s: expected %s, got %s", name, kind, actual)
		}
		// sqlite_schema stores the original declaration without its trailing ';'.
		// Compare the whole declaration so a no-op replacement trigger fails closed.
		compact := strings.Join(strings.Fields(ddl), " ")
		if !strings.Contains(strings.Join(strings.Fields(scopedWorkspaceSchemaV1), " "), compact+";") {
			return fmt.Errorf("workspace schema object %s definition mismatch", name)
		}
	}
	return nil
}
