package store

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestMessagePagesStableTiesAndDeletedCursor(t *testing.T) {
	s := compactionStore(t)
	ctx := context.Background()
	for i := 0; i < 125; i++ {
		id := fmt.Sprintf("msg_%03d", i)
		if err := s.AddMessage(ctx, id, "A", "user", id, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.DB().Exec(`update messages set created_at='2026-01-01T00:00:00Z'`); err != nil {
		t.Fatal(err)
	}
	page, err := s.PageMessages(ctx, "A", "", "", 50)
	if err != nil || len(page.Messages) != 50 || !page.HasMore || page.Messages[0].ID != "msg_075" {
		t.Fatal(page, err)
	}
	newest := page.After
	before := page.Before
	if _, err = s.DB().Exec(`delete from messages where id='msg_075'`); err != nil {
		t.Fatal(err)
	}
	older, err := s.PageMessages(ctx, "A", before, "", 50)
	if err != nil || len(older.Messages) != 50 || older.Messages[0].ID != "msg_025" || older.Messages[49].ID != "msg_074" {
		t.Fatal(older, err)
	}
	oldest, err := s.PageMessages(ctx, "A", older.Before, "", 50)
	if err != nil || len(oldest.Messages) != 25 || oldest.HasMore {
		t.Fatal(oldest, err)
	}
	for i := 125; i < 231; i++ {
		id := fmt.Sprintf("msg_%03d", i)
		if err = s.AddMessage(ctx, id, "A", "assistant", id, nil); err != nil {
			t.Fatal(err)
		}
	}
	count := 0
	for {
		page, err = s.PageMessages(ctx, "A", "", newest, 50)
		if err != nil {
			t.Fatal(err)
		}
		count += len(page.Messages)
		newest = page.After
		if !page.HasMore {
			break
		}
	}
	if count != 106 {
		t.Fatal(count)
	}
	if _, err = s.PageMessages(ctx, "foreign", before, "", 50); !errors.Is(err, ErrMessageCursor) {
		t.Fatal(err)
	}
	for _, cursor := range []string{"bad", "e30"} {
		if _, err = s.PageMessages(ctx, "A", cursor, "", 50); !errors.Is(err, ErrMessageCursor) {
			t.Fatal(err)
		}
	}
	if _, err = s.PageMessages(ctx, "A", before, newest, 50); !errors.Is(err, ErrMessageCursor) {
		t.Fatal(err)
	}
	if _, err = s.PageMessages(ctx, "A", "", "", 101); !errors.Is(err, ErrMessageCursor) {
		t.Fatal(err)
	}
}
