package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// The original turns table predates phase/claim/queue columns. Keep this
// fixture independent of initSchema, otherwise it cannot catch ordering bugs.
const legacySchema = `
create table sessions(id text primary key,parent_session_id text,title text not null default '',state_json text not null default '{}',created_at text not null,updated_at text not null,foreign key(parent_session_id) references sessions(id));
create table turns(id text primary key,session_id text not null,status text not null,prompt text not null default '',metadata_json text not null default '{}',created_at text not null,updated_at text not null,foreign key(session_id) references sessions(id) on delete cascade);
create table messages(id text primary key,session_id text not null,role text not null,content text not null default '',payload_json text not null default '{}',created_at text not null,foreign key(session_id) references sessions(id) on delete cascade);
create table turn_events(id integer primary key autoincrement,turn_id text not null,session_id text not null,seq integer not null,event_type text not null,payload_json text not null default '{}',created_at text not null,foreign key(turn_id) references turns(id) on delete cascade,foreign key(session_id) references sessions(id) on delete cascade,unique(turn_id,seq));
insert into sessions values('A',null,'legacy','{"model":"test-model"}','2026-04-01','2026-04-02');
insert into turns values('done','A','completed','original prompt','{"intent":"prompt"}','2026-04-01','2026-04-02');
insert into turns values('failed','A','failed','failed prompt','{}','2026-04-01','2026-04-02');
insert into turns values('first','A','queued','queued first','{}','2026-04-03','2026-04-03');
insert into turns values('second','A','queued','queued second','{}','2026-04-04','2026-04-04');
insert into messages values('message','A','assistant','retained answer','{"turn_id":"done"}','2026-04-02');
insert into turn_events(turn_id,session_id,seq,event_type,payload_json,created_at) values('done','A',1,'turn.completed','{"checkpoint":true}','2026-04-02');
`

func TestLegacySchemaMigratesBeforeIndexesAndPreservesHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(legacySchema); err != nil {
		t.Fatal(err)
	}
	db.Close()
	ctx := context.Background()
	var originalMessages []Message
	var originalEvents []TurnEvent
	for pass := 0; pass < 2; pass++ {
		s, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		// Identity allocation is an independent migration at routing time.
		var title, updated string
		if err := s.DB().QueryRow(`select title,updated_at from sessions where id='A'`).Scan(&title, &updated); err != nil || title != "legacy" || updated != "2026-04-02" {
			t.Fatal(title, updated, err)
		}
		phases := map[string]string{"done": "completed", "failed": "failed", "first": "queued", "second": "queued"}
		if pass == 1 {
			phases["done"] = "steered"
		}
		for id, phase := range phases {
			got, err := s.GetTurn(ctx, id)
			if err != nil || got.Phase != phase {
				t.Fatal(got, err)
			}
			if id == "done" && (got.Prompt != "original prompt" || got.UpdatedAt != "2026-04-02" || got.Metadata["intent"] != "prompt") {
				t.Fatal(got)
			}
		}
		messages, err := s.ListMessages(ctx, "A")
		if err != nil || len(messages) != 1 || messages[0].Content != "retained answer" {
			t.Fatal(messages, err)
		}
		events, err := s.ListTurnEvents(ctx, "done")
		if err != nil || len(events) != 1 || events[0].Type != "turn.completed" {
			t.Fatal(events, err)
		}
		if pass == 0 {
			originalMessages = messages
			originalEvents = events
		} else if !reflect.DeepEqual(originalMessages, messages) || !reflect.DeepEqual(originalEvents, events) {
			t.Fatal("history changed on reopen")
		}
		queued, err := s.ListQueuedTurns(ctx, "A")
		if err != nil || len(queued) != 2 || queued[0].ID != "first" || queued[1].ID != "second" {
			t.Fatal(queued, err)
		}
		var count int
		if err = s.DB().QueryRow(`select count(*) from sqlite_schema where name='idx_turns_session_phase'`).Scan(&count); err != nil || count != 1 {
			t.Fatal(count, err)
		}
		if pass == 0 {
			// A current-schema phase must not be re-derived on reopen.
			if _, err = s.DB().Exec(`update turns set phase='steered' where id='done'`); err != nil {
				t.Fatal(err)
			}
		}
		rows, err := s.DB().Query(`pragma foreign_key_check`)
		if err != nil {
			t.Fatal(err)
		}
		if rows.Next() {
			t.Fatal("foreign key violation")
		}
		rows.Close()
		if err = s.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSchemaUpgradeFailureRollsBackDDLAndBackfill(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rollback.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(legacySchema + `create table media(id integer primary key,session_id text);`); err != nil {
		t.Fatal(err)
	}
	// An incompatible legacy media table makes late index creation fail. Earlier
	// tables, ALTERs and phase backfill must not leak out of the transaction.
	if err = initSchema(db); err == nil {
		t.Fatal("expected incompatible schema error")
	}
	var n int
	if err = db.QueryRow(`select count(*) from pragma_table_info('turns') where name='phase'`).Scan(&n); err != nil || n != 0 {
		t.Fatal(n, err)
	}
	if err = db.QueryRow(`select count(*) from sqlite_schema where name='steering_queue'`).Scan(&n); err != nil || n != 0 {
		t.Fatal(n, err)
	}
	if _, err = db.Exec(`drop table media`); err != nil {
		t.Fatal(err)
	}
	if err = initSchema(db); err != nil {
		t.Fatal(err)
	}
}

func TestDuplicateColumnClassificationIsNarrow(t *testing.T) {
	for _, msg := range []string{"object already exists", "duplicate key", "duplicate value"} {
		if isDuplicateColumnError(errors.New(msg)) {
			t.Fatal(msg)
		}
	}
	if !isDuplicateColumnError(errors.New("SQL logic error: duplicate column name: phase (1)")) {
		t.Fatal("missed SQLite duplicate-column error")
	}
}

// Opt-in local-copy smoke: open a backed-up database without exposing content in
// output. The normal suite uses only synthetic fixtures above.
func TestLegacyDatabaseCopy(t *testing.T) {
	path := os.Getenv("GI_LEGACY_DB_COPY")
	if path == "" {
		t.Skip("local backup path not supplied")
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var integrity string
	if err = s.DB().QueryRow(`pragma integrity_check`).Scan(&integrity); err != nil || integrity != "ok" {
		t.Fatal(integrity, err)
	}
}
