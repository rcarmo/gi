-- Workspace index schema v1. Applied only inside the core startup transaction.
-- Legacy unscoped workspace_* tables are intentionally preserved.
-- No filesystem scan, search query API or worker is enabled by this migration.
CREATE TABLE workspace_index_workspaces (
    id INTEGER PRIMARY KEY,
    root_identity TEXT NOT NULL UNIQUE
);
CREATE TABLE workspace_index_scopes (
    workspace_id INTEGER NOT NULL REFERENCES workspace_index_workspaces(id) ON DELETE CASCADE,
    scope TEXT NOT NULL,
    config_hash TEXT NOT NULL,
    roots_json TEXT NOT NULL CHECK(json_valid(roots_json)),
    state TEXT NOT NULL DEFAULT 'never_indexed'
        CHECK(state IN ('never_indexed','indexing','ready','stale','failed')),
    committed_generation INTEGER NOT NULL DEFAULT 0,
    indexed_file_count INTEGER NOT NULL DEFAULT 0 CHECK(indexed_file_count >= 0),
    last_indexed_at_ms INTEGER,
    updated_at_ms INTEGER NOT NULL,
    last_error TEXT NOT NULL DEFAULT '',
    PRIMARY KEY(workspace_id, scope)
);
-- One writer lease per workspace, not per overlapping scope. Token changes on
-- takeover; every commit must verify owner_token and unexpired lease in its tx.
CREATE TABLE workspace_index_leases (
    workspace_id INTEGER PRIMARY KEY REFERENCES workspace_index_workspaces(id) ON DELETE CASCADE,
    owner_token TEXT NOT NULL,
    expires_at_ms INTEGER NOT NULL
);
CREATE TABLE workspace_index_documents (
    id INTEGER PRIMARY KEY,
    workspace_id INTEGER NOT NULL REFERENCES workspace_index_workspaces(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    kind TEXT NOT NULL,
    language TEXT NOT NULL DEFAULT '',
    size_bytes INTEGER NOT NULL CHECK(size_bytes >= 0),
    mtime_ns INTEGER NOT NULL,
    content_hash TEXT NOT NULL,
    chunker_version TEXT NOT NULL,
    indexed_at_ms INTEGER NOT NULL,
    UNIQUE(workspace_id, path),
    UNIQUE(workspace_id, id)
);
-- A document can belong to overlapping roots/scopes. Cleanup removes membership
-- only for the completed scan, then deletes documents with no memberships.
CREATE TABLE workspace_index_memberships (
    workspace_id INTEGER NOT NULL,
    scope TEXT NOT NULL,
    document_id INTEGER NOT NULL,
    seen_generation INTEGER NOT NULL,
    PRIMARY KEY(workspace_id, scope, document_id),
    FOREIGN KEY(workspace_id, scope) REFERENCES workspace_index_scopes(workspace_id, scope) ON DELETE CASCADE,
    FOREIGN KEY(workspace_id, document_id) REFERENCES workspace_index_documents(workspace_id, id) ON DELETE CASCADE
);
CREATE TABLE workspace_index_chunks (
    id INTEGER PRIMARY KEY,
    document_id INTEGER NOT NULL REFERENCES workspace_index_documents(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL CHECK(chunk_index >= 0),
    start_byte INTEGER NOT NULL CHECK(start_byte >= 0),
    end_byte INTEGER NOT NULL CHECK(end_byte >= start_byte),
    start_line INTEGER NOT NULL CHECK(start_line >= 1),
    end_line INTEGER NOT NULL CHECK(end_line >= start_line),
    heading TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL,
    UNIQUE(document_id, chunk_index)
);
CREATE INDEX workspace_index_chunks_by_document ON workspace_index_chunks(document_id);
CREATE INDEX workspace_index_memberships_by_document ON workspace_index_memberships(document_id);

-- FTS rowid is exactly the chunk ID. A view supplies canonical content/path;
-- triggers maintain the index for both chunk edits and document rename/language.
CREATE VIEW workspace_index_chunk_content AS
SELECT c.id, c.content, c.heading, d.path, d.language
FROM workspace_index_chunks c JOIN workspace_index_documents d ON d.id=c.document_id;
CREATE VIRTUAL TABLE workspace_index_fts USING fts5(
    content, heading, path, language UNINDEXED,
    content='workspace_index_chunk_content', content_rowid='id',
    tokenize='unicode61 remove_diacritics 2'
);
CREATE TRIGGER workspace_index_chunks_insert AFTER INSERT ON workspace_index_chunks BEGIN
    INSERT INTO workspace_index_fts(rowid,content,heading,path,language)
    SELECT id,content,heading,path,language FROM workspace_index_chunk_content WHERE id=new.id;
END;
CREATE TRIGGER workspace_index_chunks_delete BEFORE DELETE ON workspace_index_chunks BEGIN
    INSERT INTO workspace_index_fts(workspace_index_fts,rowid,content,heading,path,language)
    SELECT 'delete',id,content,heading,path,language FROM workspace_index_chunk_content WHERE id=old.id;
END;
CREATE TRIGGER workspace_index_chunks_before_update BEFORE UPDATE ON workspace_index_chunks BEGIN
    INSERT INTO workspace_index_fts(workspace_index_fts,rowid,content,heading,path,language)
    SELECT 'delete',id,content,heading,path,language FROM workspace_index_chunk_content WHERE id=old.id;
END;
CREATE TRIGGER workspace_index_chunks_after_update AFTER UPDATE ON workspace_index_chunks BEGIN
    INSERT INTO workspace_index_fts(rowid,content,heading,path,language)
    SELECT id,content,heading,path,language FROM workspace_index_chunk_content WHERE id=new.id;
END;
CREATE TRIGGER workspace_index_documents_delete BEFORE DELETE ON workspace_index_documents BEGIN
    -- Remove chunks while the parent's path/language are still available.
    DELETE FROM workspace_index_chunks WHERE document_id=old.id;
END;
CREATE TRIGGER workspace_index_documents_before_rename BEFORE UPDATE OF path,language ON workspace_index_documents BEGIN
    INSERT INTO workspace_index_fts(workspace_index_fts,rowid,content,heading,path,language)
    SELECT 'delete',id,content,heading,path,language FROM workspace_index_chunk_content WHERE id IN
        (SELECT id FROM workspace_index_chunks WHERE document_id=old.id);
END;
CREATE TRIGGER workspace_index_documents_after_rename AFTER UPDATE OF path,language ON workspace_index_documents BEGIN
    INSERT INTO workspace_index_fts(rowid,content,heading,path,language)
    SELECT id,content,heading,path,language FROM workspace_index_chunk_content WHERE id IN
        (SELECT id FROM workspace_index_chunks WHERE document_id=new.id);
END;
