package store

import "database/sql"

// UUIDs and existing HTTP cursors remain unchanged. Explicit INTEGER PRIMARY KEY
// identities survive VACUUM; AUTOINCREMENT prevents reuse after committed deletes.
// The trigger also covers transactional/legacy writers such as compaction.
func migrateMessageRows(tx *sql.Tx) error {
	for _, stmt := range []string{
		`create table if not exists message_rows (
			row_id integer primary key autoincrement,
			message_id text not null unique references messages(id) on delete cascade
		)`,
		`insert into message_rows(message_id)
		 select m.id from messages m where not exists
		 (select 1 from message_rows r where r.message_id=m.id)
		 order by m.created_at,m.id`,
		`create trigger if not exists messages_assign_row after insert on messages
		 begin insert into message_rows(message_id) values(new.id); end`,
		`create index if not exists messages_timeline on messages(session_id,created_at,id)`,
	} {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
