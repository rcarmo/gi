package store

import "database/sql"

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
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
