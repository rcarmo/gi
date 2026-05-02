package store

import "database/sql"

var Schema = []string{
	`create table if not exists workspace_documents (
		id integer primary key,
		path text not null unique,
		kind text not null,
		language text not null default '',
		size_bytes integer not null,
		mtime_ns integer not null,
		content_hash text not null,
		chunk_count integer not null default 0,
		index_state text not null default 'ready',
		last_error text not null default '',
		indexed_at_ms integer not null
	);`,
	`create table if not exists workspace_chunks (
		id integer primary key,
		document_id integer not null references workspace_documents(id) on delete cascade,
		chunk_index integer not null,
		start_byte integer not null,
		end_byte integer not null,
		start_line integer not null default 0,
		end_line integer not null default 0,
		token_estimate integer not null default 0,
		heading text not null default '',
		content text not null,
		embedding_version text not null,
		unique(document_id, chunk_index)
	);`,
	`create index if not exists idx_workspace_chunks_document_id on workspace_chunks(document_id);`,
	`create index if not exists idx_workspace_chunks_embedding_version on workspace_chunks(embedding_version);`,
	`create virtual table if not exists workspace_chunks_fts using fts5(content, heading, path, language, tokenize='unicode61');`,
	`create table if not exists workspace_index_meta (
		key text primary key,
		value text not null
	);`,
}

func InitSchema(db *sql.DB) error {
	for _, stmt := range Schema {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
