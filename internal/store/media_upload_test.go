package store

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

func TestWebUploadReuseConcurrentStoresAndReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "media.db")
	a, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	for _, id := range []string{"A", "B"} {
		if _, err = a.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	raw := bytes.Repeat([]byte("exact Unicode αβ bytes\n"), 2048)
	ids := make(chan int64, 16)
	errs := make(chan error, 16)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s := a
			if i%2 == 1 {
				s = b
			}
			m, e := s.CreateOrReuseWebMedia(ctx, "A", "file.txt", "text/plain", raw)
			if e != nil {
				errs <- e
				return
			}
			ids <- m.ID
		}(i)
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	if t.Failed() {
		return
	}
	var id int64
	for v := range ids {
		if id != 0 && v != id {
			t.Fatalf("duplicate %d != %d", v, id)
		}
		id = v
	}
	list, _ := a.ListMedia(ctx, "A")
	if len(list) != 1 {
		t.Fatal(len(list))
	}
	c, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	again, err := c.CreateOrReuseWebMedia(ctx, "A", "file.txt", "text/plain", raw)
	if err != nil || again.ID != id {
		t.Fatal(again, err)
	}
	_, content, err := c.GetMediaContent(ctx, id)
	if err != nil || !bytes.Equal(content, raw) {
		t.Fatal("bytes", err)
	}
	if !again.Compressed || again.CreatedAt == "" {
		t.Fatal("compressed metadata/timestamp missing")
	}
	for i, change := range []struct {
		session, name, mime string
		body                []byte
	}{{"B", "file.txt", "text/plain", raw}, {"A", "other.txt", "text/plain", raw}, {"A", "file.txt", "application/octet-stream", raw}, {"A", "file.txt", "text/plain", append([]byte("different"), raw...)}} {
		m, e := a.CreateOrReuseWebMedia(ctx, change.session, change.name, change.mime, change.body)
		if e != nil || m.ID == id {
			t.Fatal(i, m, e)
		}
	}
}

func TestWebUploadRejectsFalseHashAndKeepsNativeSemantics(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "media.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.CreateSession(ctx, "A", "A", nil)
	first, err := s.CreateOrReuseWebMedia(ctx, "A", "same", "text/plain", []byte("first"))
	if err != nil {
		t.Fatal(err)
	}
	// Forged metadata alone cannot make different native bytes match an upload.
	native, err := s.CreateMedia(ctx, "A", "same", "text/plain", []byte("other"), first.Metadata)
	if err != nil {
		t.Fatal(err)
	}
	if native.ID == first.ID {
		t.Fatal("native create was deduplicated")
	}
	if _, ok := native.Metadata["web_upload_sha256"]; ok {
		t.Fatal("native metadata forged web provenance")
	}
	reuse, err := s.CreateOrReuseWebMedia(ctx, "A", "same", "text/plain", []byte("first"))
	if err != nil || reuse.ID != first.ID {
		t.Fatal(reuse, err)
	}
	other, err := s.CreateOrReuseWebMedia(ctx, "A", "same", "text/plain", []byte("other"))
	if err != nil {
		t.Fatal(err)
	}
	_, raw, _ := s.GetMediaContent(ctx, other.ID)
	if string(raw) != "other" {
		t.Fatal("hash metadata trusted")
	}
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = s.CreateOrReuseWebMedia(ctx, "A", "cancelled", "text/plain", []byte("no row")); err == nil {
		t.Fatal("cancel accepted")
	}
	rows, _ := s.ListMedia(context.Background(), "A")
	for _, r := range rows {
		if r.Filename == "cancelled" {
			t.Fatal("cancel committed")
		}
	}
	if _, err = s.DB().Exec(`CREATE TRIGGER fail_media_insert BEFORE INSERT ON media BEGIN SELECT RAISE(ABORT,'fixture'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateOrReuseWebMedia(context.Background(), "A", "failed", "text/plain", []byte("rollback")); err == nil {
		t.Fatal("failure accepted")
	}
	s.DB().Exec(`DROP TRIGGER fail_media_insert`)
	if _, err = s.CreateOrReuseWebMedia(context.Background(), "A", "failed", "text/plain", []byte("rollback")); err != nil {
		t.Fatal(err)
	}
	var integrity string
	if err = s.DB().QueryRow(`PRAGMA integrity_check`).Scan(&integrity); err != nil || integrity != "ok" {
		t.Fatal(fmt.Sprint(err), integrity)
	}
}

func TestWebUploadCorruptCandidateFailsClosedAndNoLostResponseDuplicate(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "media.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.CreateSession(ctx, "A", "A", nil)
	raw := bytes.Repeat([]byte("compress me"), 200)
	first, err := s.CreateOrReuseWebMedia(ctx, "A", "file", "text/plain", raw)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a lost HTTP response by discarding the return, then retry.
	retry, err := s.CreateOrReuseWebMedia(ctx, "A", "file", "text/plain", raw)
	if err != nil || retry.ID != first.ID {
		t.Fatal(retry, err)
	}
	s.DB().Exec(`UPDATE media SET content=?,compressed=1 WHERE id=?`, []byte("not gzip"), first.ID)
	if _, err = s.CreateOrReuseWebMedia(ctx, "A", "file", "text/plain", raw); err == nil {
		t.Fatal("corrupt stored candidate silently reused")
	}
	rows, _ := s.ListMedia(ctx, "A")
	if len(rows) != 1 {
		t.Fatal("corruption silently created a second object")
	}
	// Restore bytes explicitly; no retry silently changes existing media identity.
	s.DB().Exec(`UPDATE media SET content=?,compressed=0 WHERE id=?`, raw, first.ID)
	retry, err = s.CreateOrReuseWebMedia(ctx, "A", "file", "text/plain", raw)
	if err != nil || retry.ID != first.ID {
		t.Fatal(retry, err)
	}
}

func TestWebUploadConcurrentProcesses(t *testing.T) {
	if path := os.Getenv("GI_TEST_WEB_MEDIA_DB"); path != "" {
		s, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		if _, err = s.CreateOrReuseWebMedia(context.Background(), "A", "process.txt", "text/plain", bytes.Repeat([]byte("native-process-α"), 2048)); err != nil {
			t.Fatal(err)
		}
		return
	}
	path := filepath.Join(t.TempDir(), "media.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.CreateSession(context.Background(), "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command(os.Args[0], "-test.run=^TestWebUploadConcurrentProcesses$", "-test.count=1")
			cmd.Env = append(os.Environ(), "GI_TEST_WEB_MEDIA_DB="+path)
			if out, err := cmd.CombinedOutput(); err != nil {
				errs <- fmt.Errorf("child: %w %s", err, out)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	media, err := s.ListMedia(context.Background(), "A")
	if err != nil || len(media) != 1 {
		t.Fatal(len(media), err)
	}
}
