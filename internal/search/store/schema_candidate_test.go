package store

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// Exercise FTS/membership semantics through the production migration. This is
// schema evidence, not scanner/query API or complete feature acceptance.
func TestCandidateWorkspaceSchemaFTSAndMembership(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "candidate.db")
	open := func() *sql.DB {
		t.Helper()
		db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
		if err != nil {
			t.Fatal(err)
		}
		db.SetMaxOpenConns(1)
		return db
	}
	db := open()
	defer func() { db.Close() }()
	if err := applyMigration(db); err != nil {
		t.Fatal(err)
	}
	exec := func(query string) {
		t.Helper()
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO workspace_index_workspaces VALUES(1,'/workspace/a'),(2,'/workspace/b');
 INSERT INTO workspace_index_scopes(workspace_id,scope,config_hash,roots_json,updated_at_ms) VALUES
 (1,'notes','config-a','["notes"]',1),(1,'all','config-a','["notes",".pi/skills"]',1),(2,'all','config-b','["notes"]',1);
 INSERT INTO workspace_index_documents(id,workspace_id,path,kind,size_bytes,mtime_ns,content_hash,chunker_version,indexed_at_ms)
 VALUES(10,1,'notes/orchid.md','text',12,1,'hash','lines-v1',1),(20,2,'notes/orchid.md','text',12,1,'hash','lines-v1',1);
 INSERT INTO workspace_index_memberships VALUES(1,'notes',10,1),(1,'all',10,1),(2,'all',20,1);
 INSERT INTO workspace_index_chunks VALUES(100,10,0,0,12,1,1,'Flowers','orchid violet'),(200,20,0,0,12,1,1,'Flowers','orchid remote');`)
	hits := func(term string) int {
		t.Helper()
		var n int
		if err := db.QueryRow("select count(*) from workspace_index_fts where workspace_index_fts match ?", term).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if hits("orchid") != 2 {
		t.Fatal("missing inserts")
	}
	var scoped int
	if err := db.QueryRow(`SELECT count(*) FROM workspace_index_fts f JOIN workspace_index_chunks c ON c.id=f.rowid
 JOIN workspace_index_memberships m ON m.document_id=c.document_id
 WHERE workspace_index_fts MATCH 'orchid' AND m.workspace_id=1 AND m.scope='notes'`).Scan(&scoped); err != nil || scoped != 1 {
		t.Fatal(scoped, err)
	}
	if _, err := db.Exec("INSERT INTO workspace_index_memberships VALUES(1,'notes',20,1)"); err == nil {
		t.Fatal("foreign workspace membership accepted")
	}
	exec("UPDATE workspace_index_chunks SET content='daisy violet' WHERE id=100")
	if hits("orchid") != 2 || hits("daisy") != 1 {
		t.Fatal("content update lost path token or new content")
	}
	exec("UPDATE workspace_index_documents SET path='notes/daisy.md' WHERE id=10")
	if hits("orchid") != 1 {
		t.Fatal("old path still indexed")
	}
	// Removing one overlapping membership must not remove shared content.
	exec("DELETE FROM workspace_index_memberships WHERE workspace_id=1 AND scope='notes'; DELETE FROM workspace_index_documents WHERE workspace_id=1 AND NOT EXISTS(SELECT 1 FROM workspace_index_memberships m WHERE m.document_id=workspace_index_documents.id)")
	if hits("daisy") != 1 {
		t.Fatal("scoped cleanup deleted shared document")
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec("DELETE FROM workspace_index_documents WHERE id=10"); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if hits("daisy") != 1 {
		t.Fatal("rollback lost FTS")
	}
	exec("DELETE FROM workspace_index_documents WHERE id=10")
	if hits("daisy") != 0 || hits("orchid") != 1 {
		t.Fatal("cascade left stale FTS")
	}
	exec("INSERT INTO workspace_index_fts(workspace_index_fts,rank) VALUES('integrity-check',1)")
	exec("INSERT INTO workspace_index_leases VALUES(1,'owner-a',100)")
	if _, err := db.Exec("INSERT INTO workspace_index_leases VALUES(1,'owner-b',200)"); err == nil {
		t.Fatal("second owner accepted")
	}
	result, err := db.Exec("UPDATE workspace_index_leases SET owner_token='owner-b',expires_at_ms=300 WHERE workspace_id=1 AND expires_at_ms<=200")
	if err != nil {
		t.Fatal(err)
	}
	if n, _ := result.RowsAffected(); n != 1 {
		t.Fatal("lease takeover failed")
	}
	result, err = db.Exec("UPDATE workspace_index_leases SET expires_at_ms=400 WHERE workspace_id=1 AND owner_token='owner-a'")
	if err != nil {
		t.Fatal(err)
	}
	if n, _ := result.RowsAffected(); n != 0 {
		t.Fatal("stale owner renewed")
	}
	exec("UPDATE workspace_index_scopes SET state='failed',last_error='fixture failure',committed_generation=3,indexed_file_count=1,last_indexed_at_ms=10 WHERE workspace_id=2 AND scope='all'")
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = open()
	if hits("orchid") != 1 {
		t.Fatal("reopen lost index")
	}
	var state string
	var generation int
	if err := db.QueryRow("SELECT state,committed_generation FROM workspace_index_scopes WHERE workspace_id=2 AND scope='all'").Scan(&state, &generation); err != nil || state != "failed" || generation != 3 {
		t.Fatal(state, generation, err)
	}
	exec("INSERT INTO workspace_index_fts(workspace_index_fts) VALUES('rebuild'); INSERT INTO workspace_index_fts(workspace_index_fts,rank) VALUES('integrity-check',1)")
	if hits("orchid") != 1 {
		t.Fatal("FTS rebuild changed content")
	}
	exec("DELETE FROM workspace_index_workspaces WHERE id=2")
	if hits("orchid") != 0 {
		t.Fatal("workspace deletion left FTS rows")
	}
	exec("INSERT INTO workspace_index_fts(workspace_index_fts,rank) VALUES('integrity-check',1)")
}
