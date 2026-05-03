package store

import (
	"database/sql"
	"fmt"
	"strings"
)

func initSchema(db *sql.DB) error {
	stmts := []string{
		`create table if not exists sessions (
			id text primary key,
			parent_session_id text,
			title text not null default '',
			state_json text not null default '{}',
			created_at text not null,
			updated_at text not null,
			foreign key(parent_session_id) references sessions(id)
		);`,
		`create index if not exists idx_sessions_updated_at on sessions(updated_at desc);`,
		`create index if not exists idx_sessions_parent on sessions(parent_session_id);`,
		`create index if not exists idx_sessions_state_active_turn on sessions(json_extract(state_json, '$.active_turn_id'));`,
		`create index if not exists idx_sessions_state_model on sessions(json_extract(state_json, '$.model'));`,

		`create table if not exists messages (
			id text primary key,
			session_id text not null,
			role text not null,
			content text not null default '',
			payload_json text not null default '{}',
			created_at text not null,
			foreign key(session_id) references sessions(id) on delete cascade
		);`,
		`create index if not exists idx_messages_session_created on messages(session_id, created_at asc);`,
		`create index if not exists idx_messages_payload_kind on messages(json_extract(payload_json, '$.kind'));`,
		`create index if not exists idx_messages_payload_intent on messages(json_extract(payload_json, '$.intent'));`,

		`create table if not exists turns (
			id text primary key,
			session_id text not null,
			status text not null,
			prompt text not null default '',
			metadata_json text not null default '{}',
			created_at text not null,
			updated_at text not null,
			foreign key(session_id) references sessions(id) on delete cascade
		);`,
		`create index if not exists idx_turns_session_status on turns(session_id, status, updated_at desc);`,
		`create index if not exists idx_turns_metadata_intent on turns(json_extract(metadata_json, '$.intent'));`,

		`create table if not exists turn_events (
			id integer primary key autoincrement,
			turn_id text not null,
			session_id text not null,
			seq integer not null,
			event_type text not null,
			payload_json text not null default '{}',
			created_at text not null,
			foreign key(turn_id) references turns(id) on delete cascade,
			foreign key(session_id) references sessions(id) on delete cascade,
			unique(turn_id, seq)
		);`,
		`create index if not exists idx_turn_events_turn_seq on turn_events(turn_id, seq asc);`,
		`create index if not exists idx_turn_events_session_created on turn_events(session_id, created_at asc);`,
		`create index if not exists idx_turn_events_type on turn_events(event_type, created_at asc);`,
		`create index if not exists idx_turn_events_phase on turn_events(json_extract(payload_json, '$.phase'));`,
		`create index if not exists idx_turn_events_checkpoint on turn_events(json_extract(payload_json, '$.checkpoint'));`,

		`create table if not exists media (
			id integer primary key autoincrement,
			session_id text not null,
			filename text not null default '',
			content_type text,
			metadata_json text not null default '{}',
			original_size integer not null default 0,
			compressed_size integer not null default 0,
			compressed integer not null default 0,
			content blob not null,
			created_at text not null,
			updated_at text not null,
			foreign key(session_id) references sessions(id) on delete cascade
		);`,
		`create index if not exists idx_media_session on media(session_id, created_at desc);`,
		`create index if not exists idx_media_filename on media(filename);`,
		`create table if not exists vfs_files (
			namespace text not null,
			path text not null,
			content_type text,
			metadata_json text not null default '{}',
			original_size integer not null default 0,
			compressed_size integer not null default 0,
			compressed integer not null default 0,
			content blob not null,
			created_at text not null,
			updated_at text not null,
			primary key (namespace, path)
		);`,
		`create index if not exists idx_vfs_namespace_path on vfs_files(namespace, path);`,
		`create table if not exists kv_store (
			namespace text not null,
			key text not null,
			value blob not null,
			created_at text not null,
			updated_at text not null,
			primary key (namespace, key)
		);`,
		`create index if not exists idx_kv_store_namespace on kv_store(namespace, updated_at desc);`,
		`create table if not exists routing_events (
			id integer primary key autoincrement,
			turn_id text,
			source_session_id text not null,
			target_session_id text,
			source_agent_id text,
			target_agent_id text not null,
			mode text not null,
			matched_by text,
			routing_policy text,
			requested_agent_id text,
			metadata_json text not null default '{}',
			created_at text not null,
			foreign key(turn_id) references turns(id) on delete cascade,
			foreign key(source_session_id) references sessions(id) on delete cascade,
			foreign key(target_session_id) references sessions(id) on delete set null
		);`,
		`create index if not exists idx_routing_events_source on routing_events(source_session_id, created_at desc);`,
		`create index if not exists idx_routing_events_target on routing_events(target_session_id, created_at desc);`,
		`create index if not exists idx_routing_events_turn on routing_events(turn_id);`,
		`create index if not exists idx_routing_events_mode on routing_events(mode);`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	for _, alter := range []string{
		`alter table sessions add column scope_json text not null default '{}'`,
		`alter table sessions add column aliases_json text not null default '[]'`,
	} {
		if _, err := db.Exec(alter); err != nil && !isDuplicateColumnError(err) {
			return err
		}
	}
	return nil
}

func isDuplicateColumnError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate column") || strings.Contains(msg, "already exists") || strings.Contains(msg, "duplicate") || strings.Contains(msg, fmt.Sprintf("%q", "scope_json"))
}
