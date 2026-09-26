package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const legacyMessageRowsSchema = `
create table sessions(
	id text primary key,
	parent_session_id text,
	title text not null default '',
	state_json text not null default '{}',
	created_at text not null,
	updated_at text not null,
	foreign key(parent_session_id) references sessions(id)
);
create table messages(
	id text primary key,
	session_id text not null,
	role text not null,
	content text not null default '',
	payload_json text not null default '{}',
	created_at text not null,
	foreign key(session_id) references sessions(id) on delete cascade
);
`

type messageRowRef struct {
	RowID     int64
	MessageID string
}

func legacyMessageRowsFixture(t *testing.T, name, setup string) (string, [][]any) {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	db := openSQLiteDB(t, path)
	defer db.Close()
	if _, err := db.Exec(`pragma foreign_keys=on`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(legacyMessageRowsSchema + setup); err != nil {
		t.Fatal(err)
	}
	return path, messageSnapshot(t, db)
}

func messageSnapshot(t *testing.T, db *sql.DB) [][]any {
	t.Helper()
	rows, err := db.Query(`select id,session_id,role,content,payload_json,created_at from messages order by id asc`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var snapshot [][]any
	for rows.Next() {
		values := make([]any, 6)
		ptrs := make([]any, len(values))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatal(err)
		}
		for i, v := range values {
			if raw, ok := v.([]byte); ok {
				values[i] = append([]byte(nil), raw...)
			}
		}
		snapshot = append(snapshot, values)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func openSQLiteDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func messageRowsByRowID(t *testing.T, db *sql.DB) []messageRowRef {
	t.Helper()
	rows, err := db.Query(`select row_id,message_id from message_rows order by row_id asc`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []messageRowRef
	for rows.Next() {
		var item messageRowRef
		if err := rows.Scan(&item.RowID, &item.MessageID); err != nil {
			t.Fatal(err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func messageRowID(t *testing.T, db *sql.DB, messageID string) int64 {
	t.Helper()
	var rowID int64
	if err := db.QueryRow(`select row_id from message_rows where message_id=?`, messageID).Scan(&rowID); err != nil {
		t.Fatal(err)
	}
	return rowID
}

func sqliteSequence(t *testing.T, db *sql.DB, table string) int64 {
	t.Helper()
	var seq int64
	if err := db.QueryRow(`select seq from sqlite_sequence where name=?`, table).Scan(&seq); err != nil {
		t.Fatal(err)
	}
	return seq
}

func countQuery(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func messageIDs(messages []Message) []string {
	ids := make([]string, len(messages))
	for i, message := range messages {
		ids[i] = message.ID
	}
	return ids
}

func requireNoForeignKeyViolations(t *testing.T, db *sql.DB) {
	t.Helper()
	rows, err := db.Query(`pragma foreign_key_check`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if rows.Next() {
		t.Fatal("foreign key violation")
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestMessageRowsLegacyBackfillStableAcrossReopenAndVacuum(t *testing.T) {
	path, beforeMessages := legacyMessageRowsFixture(t, "legacy-message-rows.db", `
insert into sessions values('A',null,'legacy','{"model":"test-model"}','2026-05-01','2026-05-01');
insert into messages values('99999999-9999-9999-9999-999999999999','A','assistant','latest','{"uuid":"99999999-9999-9999-9999-999999999999","kind":"assistant","http":{"list":true}}','2026-05-03T00:00:00Z');
insert into messages values('33333333-3333-3333-3333-333333333333','A','assistant','tie later id','{"uuid":"33333333-3333-3333-3333-333333333333","kind":"assistant"}','2026-05-02T00:00:00Z');
insert into messages values('11111111-1111-1111-1111-111111111111','A','user','earliest','{"uuid":"11111111-1111-1111-1111-111111111111","kind":"user"}','2026-05-01T00:00:00Z');
insert into messages values('22222222-2222-2222-2222-222222222222','A','user','tie earlier id','{"uuid":"22222222-2222-2222-2222-222222222222","nested":{"kept":true}}','2026-05-02T00:00:00Z');
`)
	expectedRows := []messageRowRef{
		{1, "11111111-1111-1111-1111-111111111111"},
		{2, "22222222-2222-2222-2222-222222222222"},
		{3, "33333333-3333-3333-3333-333333333333"},
		{4, "99999999-9999-9999-9999-999999999999"},
	}
	ctx := context.Background()
	for pass := 0; pass < 2; pass++ {
		s, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(beforeMessages, messageSnapshot(t, s.DB())) {
			s.Close()
			t.Fatal("messages changed during backfill/reopen")
		}
		if got := messageRowsByRowID(t, s.DB()); !reflect.DeepEqual(got, expectedRows) {
			s.Close()
			t.Fatal(got)
		}
		if seq := sqliteSequence(t, s.DB(), "message_rows"); seq != 4 {
			s.Close()
			t.Fatal(seq)
		}
		messages, err := s.ListMessages(ctx, "A")
		if err != nil {
			s.Close()
			t.Fatal(err)
		}
		if got := messageIDs(messages); !reflect.DeepEqual(got, []string{
			"11111111-1111-1111-1111-111111111111",
			"22222222-2222-2222-2222-222222222222",
			"33333333-3333-3333-3333-333333333333",
			"99999999-9999-9999-9999-999999999999",
		}) {
			s.Close()
			t.Fatal(got)
		}
		nested, ok := messages[1].Payload["nested"].(map[string]any)
		if !ok || messages[1].Payload["uuid"] != "22222222-2222-2222-2222-222222222222" || nested["kept"] != true {
			s.Close()
			t.Fatal(messages[1].Payload)
		}
		requireNoForeignKeyViolations(t, s.DB())
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}
	raw := openSQLiteDB(t, path)
	if _, err := raw.Exec(`vacuum`); err != nil {
		raw.Close()
		t.Fatal(err)
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if got := messageRowsByRowID(t, s.DB()); !reflect.DeepEqual(got, expectedRows) {
		t.Fatal(got)
	}
	if seq := sqliteSequence(t, s.DB(), "message_rows"); seq != 4 {
		t.Fatal(seq)
	}
	if _, err := s.DB().Exec(`insert into messages(id,session_id,role,content,payload_json,created_at) values(?,?,?,?,?,?)`,
		"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "A", "assistant", "after vacuum", `{"uuid":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}`, "2026-05-04T00:00:00Z",
	); err != nil {
		t.Fatal(err)
	}
	if rowID := messageRowID(t, s.DB(), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"); rowID != 5 {
		t.Fatal(rowID)
	}
	messages, err := s.ListMessages(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	if messages[len(messages)-1].ID != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatal(messages)
	}
}

func TestMessageRowsNeverReuseIDsAndCascadeDeletes(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "message-rows-sequence.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	for _, id := range []string{"A", "B"} {
		if _, err := s.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	for _, item := range []struct {
		id, sessionID, role string
	}{
		{"a1", "A", "user"},
		{"a2", "A", "assistant"},
		{"b1", "B", "user"},
	} {
		if err := s.AddMessage(ctx, item.id, item.sessionID, item.role, item.id, nil); err != nil {
			t.Fatal(err)
		}
	}
	if got := messageRowsByRowID(t, s.DB()); !reflect.DeepEqual(got, []messageRowRef{{1, "a1"}, {2, "a2"}, {3, "b1"}}) {
		t.Fatal(got)
	}
	if _, err := s.DB().Exec(`delete from messages where id='b1'`); err != nil {
		t.Fatal(err)
	}
	if count := countQuery(t, s.DB(), `select count(*) from message_rows where message_id='b1'`); count != 0 {
		t.Fatal(count)
	}
	if err := s.AddMessage(ctx, "b2", "B", "assistant", "b2", nil); err != nil {
		t.Fatal(err)
	}
	if rowID := messageRowID(t, s.DB(), "b2"); rowID != 4 {
		t.Fatal(rowID)
	}
	if _, err := s.DB().Exec(`delete from sessions where id='A'`); err != nil {
		t.Fatal(err)
	}
	if count := countQuery(t, s.DB(), `select count(*) from message_rows where message_id in ('a1','a2')`); count != 0 {
		t.Fatal(count)
	}
	if got := messageRowsByRowID(t, s.DB()); !reflect.DeepEqual(got, []messageRowRef{{4, "b2"}}) {
		t.Fatal(got)
	}
	if _, err := s.DB().Exec(`delete from messages`); err != nil {
		t.Fatal(err)
	}
	if count := countQuery(t, s.DB(), `select count(*) from message_rows`); count != 0 {
		t.Fatal(count)
	}
	if err := s.AddMessage(ctx, "b3", "B", "assistant", "b3", nil); err != nil {
		t.Fatal(err)
	}
	if rowID := messageRowID(t, s.DB(), "b3"); rowID != 5 {
		t.Fatal(rowID)
	}
	if seq := sqliteSequence(t, s.DB(), "message_rows"); seq != 5 {
		t.Fatal(seq)
	}
	requireNoForeignKeyViolations(t, s.DB())
}

func TestMessageRowsTriggerFailureRollsBackMessageInsert(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "message-rows-trigger.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(`
create trigger fail_message_rows before insert on message_rows
when new.message_id in ('fail_api','fail_tx')
begin
	select raise(abort,'message row rejected');
end;`); err != nil {
		t.Fatal(err)
	}
	if err := s.AddMessage(ctx, "fail_api", "A", "user", "boom", map[string]any{"kind": "test"}); err == nil || !strings.Contains(err.Error(), "message row rejected") {
		t.Fatal(err)
	}
	if count := countQuery(t, s.DB(), `select count(*) from messages where id='fail_api'`); count != 0 {
		t.Fatal(count)
	}
	if count := countQuery(t, s.DB(), `select count(*) from message_rows where message_id='fail_api'`); count != 0 {
		t.Fatal(count)
	}
	tx, err := s.DB().BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `insert into messages(id,session_id,role,content,payload_json,created_at) values('fail_tx','A','assistant','boom','{}','2026-06-01T00:00:00Z')`); err == nil || !strings.Contains(err.Error(), "message row rejected") {
		tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if count := countQuery(t, s.DB(), `select count(*) from messages where id='fail_tx'`); count != 0 {
		t.Fatal(count)
	}
	if count := countQuery(t, s.DB(), `select count(*) from message_rows where message_id='fail_tx'`); count != 0 {
		t.Fatal(count)
	}
	if err := s.AddMessage(ctx, "ok", "A", "assistant", "ok", nil); err != nil {
		t.Fatal(err)
	}
	if got := messageRowsByRowID(t, s.DB()); !reflect.DeepEqual(got, []messageRowRef{{1, "ok"}}) {
		t.Fatal(got)
	}
}

func TestMessageRowsMigrationFailureRollsBackUpgrade(t *testing.T) {
	path, beforeMessages := legacyMessageRowsFixture(t, "message-rows-rollback.db", `
insert into sessions values('A',null,'legacy','{"model":"test-model"}','2026-05-01','2026-05-01');
insert into messages values('legacy-fail','A','assistant','legacy','{"uuid":"legacy-fail"}','2026-05-01T00:00:00Z');
create table message_rows(
	row_id integer primary key autoincrement,
	message_id text not null unique references messages(id) on delete cascade
);
create trigger fail_backfill before insert on message_rows when new.message_id='legacy-fail' begin select raise(abort,'backfill rejected'); end;
`)
	if s, err := Open(path); err == nil {
		s.Close()
		t.Fatal("expected migration failure")
	}
	db := openSQLiteDB(t, path)
	if count := countQuery(t, db, `select count(*) from message_rows`); count != 0 {
		db.Close()
		t.Fatal(count)
	}
	if count := countQuery(t, db, `select count(*) from sqlite_schema where type='trigger' and name='messages_assign_row'`); count != 0 {
		db.Close()
		t.Fatal(count)
	}
	if count := countQuery(t, db, `select count(*) from pragma_table_info('sessions') where name='aliases_json'`); count != 0 {
		db.Close()
		t.Fatal("partial alter leaked", count)
	}
	if count := countQuery(t, db, `select count(*) from sqlite_schema where name='steering_queue'`); count != 0 {
		db.Close()
		t.Fatal("partial table leaked", count)
	}
	if !reflect.DeepEqual(beforeMessages, messageSnapshot(t, db)) {
		db.Close()
		t.Fatal("messages changed after rollback")
	}
	if _, err := db.Exec(`drop trigger fail_backfill`); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if got := messageRowsByRowID(t, s.DB()); !reflect.DeepEqual(got, []messageRowRef{{1, "legacy-fail"}}) {
		t.Fatal(got)
	}
}
