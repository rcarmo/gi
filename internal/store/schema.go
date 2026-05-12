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

		`create table if not exists session_identities (
			session_id text primary key,
			agent_id text not null,
			channel text not null,
			account text not null,
			scope_version integer not null default 1,
			canonical_scope_signature text not null unique,
			opaque_session_key text not null unique,
			is_main_session integer not null default 0,
			created_at text not null,
			updated_at text not null,
			foreign key(session_id) references sessions(id) on delete cascade
		);`,
		`create index if not exists idx_session_identities_agent on session_identities(agent_id, updated_at desc);`,
		`create index if not exists idx_session_identities_channel_account on session_identities(channel, account, updated_at desc);`,

		`create table if not exists session_identity_dimensions (
			session_id text not null,
			dimension_name text not null,
			dimension_value text not null,
			ordinal integer not null,
			primary key (session_id, dimension_name),
			foreign key(session_id) references sessions(id) on delete cascade
		);`,
		`create index if not exists idx_sid_dim_lookup on session_identity_dimensions(dimension_name, dimension_value);`,
		`create index if not exists idx_sid_dim_session_ordinal on session_identity_dimensions(session_id, ordinal);`,

		`create table if not exists session_aliases (
			alias text primary key,
			session_id text not null,
			alias_kind text not null,
			created_at text not null,
			updated_at text not null,
			foreign key(session_id) references sessions(id) on delete cascade
		);`,
		`create index if not exists idx_session_aliases_session on session_aliases(session_id);`,

		`create table if not exists session_channel_bindings (
			id integer primary key autoincrement,
			session_id text not null,
			channel text not null,
			account text not null,
			binding_type text not null,
			remote_identity text not null,
			metadata_json text not null default '{}',
			created_at text not null,
			updated_at text not null,
			foreign key(session_id) references sessions(id) on delete cascade,
			unique(channel, account, remote_identity)
		);`,
		`create index if not exists idx_session_channel_bindings_session on session_channel_bindings(session_id, channel, account);`,

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
			phase text not null default 'queued',
			prompt text not null default '',
			metadata_json text not null default '{}',
			claimed_by text,
			claimed_at text,
			started_at text,
			finished_at text,
			created_at text not null,
			updated_at text not null,
			foreign key(session_id) references sessions(id) on delete cascade
		);`,
		`create index if not exists idx_turns_session_status on turns(session_id, status, updated_at desc);`,
		`create index if not exists idx_turns_session_phase on turns(session_id, phase, updated_at desc);`,
		`create index if not exists idx_turns_metadata_intent on turns(json_extract(metadata_json, '$.intent'));`,

		`create table if not exists subturns (
			id integer primary key autoincrement,
			parent_turn_id text not null,
			parent_session_id text not null,
			child_turn_id text not null unique,
			child_session_id text not null,
			delivery_mode text not null default 'sync',
			status text not null default 'running',
			depth integer not null default 1,
			metadata_json text not null default '{}',
			created_at text not null,
			updated_at text not null,
			finished_at text,
			foreign key(parent_turn_id) references turns(id) on delete cascade,
			foreign key(child_turn_id) references turns(id) on delete cascade,
			foreign key(parent_session_id) references sessions(id) on delete cascade,
			foreign key(child_session_id) references sessions(id) on delete cascade,
			unique(parent_turn_id, child_turn_id)
		);`,
		`create index if not exists idx_subturns_parent on subturns(parent_turn_id, created_at asc);`,
		`create index if not exists idx_subturns_child on subturns(child_turn_id);`,
		`create index if not exists idx_subturns_parent_session on subturns(parent_session_id, created_at asc);`,
		`create index if not exists idx_subturns_child_session on subturns(child_session_id, created_at asc);`,

		`create table if not exists session_active_turns (
			session_id text primary key,
			turn_id text not null,
			worker_id text,
			claim_token text not null,
			claimed_at text not null,
			updated_at text not null,
			foreign key(session_id) references sessions(id) on delete cascade,
			foreign key(turn_id) references turns(id) on delete cascade
		);`,
		`create index if not exists idx_session_active_turns_turn on session_active_turns(turn_id);`,

		`create table if not exists turn_failures (
			turn_id text primary key,
			session_id text not null,
			failure_kind text not null,
			hold_state text not null,
			summary text not null default '',
			resolution_state text not null default '',
			resolution_summary text not null default '',
			resolved_at text,
			resolved_turn_id text,
			created_at text not null,
			updated_at text not null,
			foreign key(turn_id) references turns(id) on delete cascade,
			foreign key(session_id) references sessions(id) on delete cascade,
			foreign key(resolved_turn_id) references turns(id) on delete set null
		);`,
		`create index if not exists idx_turn_failures_session on turn_failures(session_id, updated_at desc);`,
		`create index if not exists idx_turn_failures_kind on turn_failures(failure_kind, updated_at desc);`,

		`create table if not exists steering_queue (
			id integer primary key autoincrement,
			session_id text not null,
			turn_id text,
			role text not null default 'user',
			content text not null default '',
			payload_json text not null default '{}',
			media_json text not null default '[]',
			queue_mode text not null default 'one-at-a-time',
			status text not null default 'queued',
			created_at text not null,
			updated_at text not null,
			foreign key(session_id) references sessions(id) on delete cascade,
			foreign key(turn_id) references turns(id) on delete set null
		);`,
		`create index if not exists idx_steering_queue_session_status on steering_queue(session_id, status, id);`,

		`create table if not exists inbound_work_queue (
			id integer primary key autoincrement,
			source_kind text not null,
			session_id text,
			explicit_session_key text not null default '',
			envelope_json text not null default '{}',
			status text not null default 'queued',
			attempt_count integer not null default 0,
			last_error text not null default '',
			next_attempt_at text,
			claimed_by text,
			claimed_at text,
			created_at text not null,
			updated_at text not null,
			foreign key(session_id) references sessions(id) on delete set null
		);`,
		`create index if not exists idx_inbound_work_queue_status on inbound_work_queue(status, id);`,
		`create index if not exists idx_inbound_work_queue_session on inbound_work_queue(session_id, status, id);`,

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
		`create table if not exists hook_invocations (
			id integer primary key autoincrement,
			turn_id text,
			session_id text,
			hook_name text not null,
			hook_phase text not null,
			hook_source text not null,
			action text not null default 'continue',
			request_json text not null default '{}',
			response_json text not null default '{}',
			error_text text not null default '',
			duration_ms integer not null default 0,
			created_at text not null,
			foreign key(turn_id) references turns(id) on delete cascade,
			foreign key(session_id) references sessions(id) on delete cascade
		);`,
		`create index if not exists idx_hook_invocations_turn on hook_invocations(turn_id, id);`,
		`create index if not exists idx_hook_invocations_phase on hook_invocations(hook_phase, created_at);`,
		`create index if not exists idx_hook_invocations_session on hook_invocations(session_id, created_at);`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	for _, alter := range []string{
		`alter table sessions add column scope_json text not null default '{}'`,
		`alter table sessions add column aliases_json text not null default '[]'`,
		`alter table turns add column phase text not null default 'queued'`,
		`alter table turns add column claimed_by text`,
		`alter table turns add column claimed_at text`,
		`alter table turns add column started_at text`,
		`alter table turns add column finished_at text`,
		`alter table turn_failures add column resolution_state text not null default ''`,
		`alter table turn_failures add column resolution_summary text not null default ''`,
		`alter table turn_failures add column resolved_at text`,
		`alter table turn_failures add column resolved_turn_id text`,
		`alter table inbound_work_queue add column attempt_count integer not null default 0`,
		`alter table inbound_work_queue add column last_error text not null default ''`,
		`alter table inbound_work_queue add column next_attempt_at text`,
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
