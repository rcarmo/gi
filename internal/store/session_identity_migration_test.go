package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	gisession "github.com/rcarmo/gi/internal/session"
)

func legacyIdentityFixture(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "legacy-identities.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	schema := strings.Replace(legacySchema, "state_json text not null default '{}',created_at", "state_json text not null default '{}',scope_json text not null default '{}',aliases_json text not null default '[]',created_at", 1)
	schema = strings.Replace(schema, "insert into sessions values('A',null,'legacy','{\"model\":\"test-model\"}',", "insert into sessions(id,parent_session_id,title,state_json,created_at,updated_at) values('A',null,'legacy','{\"model\":\"test-model\"}',", 1)
	if _, err = db.Exec(schema); err != nil {
		t.Fatal(err)
	}
	// Capture the original allocation representation, including its aliases.
	if _, err = db.Exec(`insert into sessions(id,title,state_json,scope_json,aliases_json,created_at,updated_at) values('scoped','@agent','{}',?,?, '2026-04-01','2026-04-02')`, `{"version":1,"agent_id":"agent","channel":"gi","account":"default","dimensions":["chat"],"values":{"chat":"direct:scoped"}}`, `["agent:agent:gi:chat:scoped","gi:scoped"]`); err != nil {
		t.Fatal(err)
	}
	return path
}
func TestLegacySessionIdentityMigrationPreservesHistoryAndReopens(t *testing.T) {
	path := legacyIdentityFixture(t)
	var baseline []Message
	for i := 0; i < 2; i++ {
		s, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		sessions, err := s.ListSessions(t.Context())
		if err != nil || len(sessions) != 2 {
			t.Fatal(sessions, err)
		}
		legacy, err := s.GetSession(t.Context(), "A")
		if err != nil || legacy.Scope == nil || legacy.Scope.Values["chat"] != "direct:a" || legacy.Title != "legacy" || legacy.UpdatedAt != "2026-04-02" {
			t.Fatal(legacy, err)
		}
		scoped, err := s.GetSession(t.Context(), "scoped")
		if err != nil || scoped.Scope.AgentID != "agent" || scoped.Scope.Values["chat"] != "direct:scoped" || len(scoped.Aliases) != 2 {
			t.Fatal(scoped, err)
		}
		id, err := s.GetSessionIdentity(t.Context(), "scoped")
		if err != nil || id.OpaqueSessionKey != gisession.BuildSessionKey(*scoped.Scope) {
			t.Fatal(id, err)
		}
		resolved, err := s.ResolveSessionIDByAlias(t.Context(), "gi:scoped")
		if err != nil || resolved != "scoped" {
			t.Fatal(resolved, err)
		}
		resolved, err = s.FindSessionByAllocation(t.Context(), gisession.AllocateDefaultSession("agent", "gi", "default", "scoped"))
		if err != nil || resolved != "scoped" {
			t.Fatal(resolved, err)
		}
		var n int
		if err = s.DB().QueryRow(`select count(*) from session_identities`).Scan(&n); err != nil || n != 2 {
			t.Fatal(n, err)
		}
		msgs, err := s.ListMessages(t.Context(), "A")
		if err != nil || len(msgs) != 1 {
			t.Fatal(msgs, err)
		}
		if i == 0 {
			baseline = msgs
		} else if !reflect.DeepEqual(baseline, msgs) {
			t.Fatal("history changed", msgs)
		}
		var title, updated string
		if err = s.DB().QueryRow(`select title,updated_at from sessions where id='A'`).Scan(&title, &updated); err != nil || title != "legacy" || updated != "2026-04-02" {
			t.Fatal(title, updated, err)
		}
		if err = s.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
func TestLegacySessionIdentityMigrationRollsBackLateFailure(t *testing.T) {
	path := legacyIdentityFixture(t)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`create table session_identities(session_id text primary key,agent_id text not null,channel text not null,account text not null,scope_version integer not null default 1,canonical_scope_signature text not null unique,opaque_session_key text not null unique,is_main_session integer not null default 0,created_at text not null,updated_at text not null);create trigger fail_scoped before insert on session_identities when new.session_id='scoped' begin select raise(abort,'late failure');end;`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if s, err := Open(path); err == nil {
		s.Close()
		t.Fatal("late failure accepted")
	}
	db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	if err = db.QueryRow(`select count(*) from session_identities`).Scan(&n); err != nil || n != 0 {
		t.Fatal("partial identities", n, err)
	}
	if err = db.QueryRow(`select count(*) from pragma_table_info('turns') where name='phase'`).Scan(&n); err != nil || n != 0 {
		t.Fatal("partial DDL", n, err)
	}
	if _, err = db.Exec(`drop trigger fail_scoped`); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.ListSessions(context.Background()); err != nil {
		t.Fatal(err)
	}
}
func TestLegacySessionIdentityMigrationRejectsIncompleteOrConflictingScope(t *testing.T) {
	for _, tc := range []struct{ name, scope string }{
		{"incomplete", `{"version":1,"agent_id":"agent","channel":"gi","account":"default","dimensions":["chat"],"values":{}}`},
		{"collision", `{"version":1,"agent_id":"gi","channel":"gi","account":"default","dimensions":["chat"],"values":{"chat":"direct:A"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := legacyIdentityFixture(t)
			db, err := sql.Open("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = db.Exec(`update sessions set scope_json=? where id='scoped'`, tc.scope); err != nil {
				t.Fatal(err)
			}
			db.Close()
			if s, err := Open(path); err == nil {
				s.Close()
				t.Fatal("invalid scope accepted")
			}
			db, err = sql.Open("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			var count int
			if err = db.QueryRow(`select count(*) from sqlite_schema where name='session_identities'`).Scan(&count); err != nil || count != 0 {
				t.Fatal("partial schema migration", count, err)
			}
		})
	}
}

func TestLegacyColumnDoesNotRepairPostAllocationMissingIdentity(t *testing.T) {
	path := legacyIdentityFixture(t)
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(`delete from session_identities where session_id='scoped'`); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.GetSession(t.Context(), "scoped"); err == nil {
		t.Fatal("repaired a post-allocation missing identity")
	}
}

func TestCurrentSchemaMissingIdentityDoesNotBackfill(t *testing.T) {
	path := filepath.Join(t.TempDir(), "current.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateSession(t.Context(), "current", "current", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.DB().Exec(`delete from session_identities where session_id='current'`); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.GetSession(t.Context(), "current"); err == nil {
		t.Fatal("synthesised identity for current-schema row")
	}
}
