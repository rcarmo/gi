package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func enrolledManager(t *testing.T) (*Manager, string) {
	t.Helper()
	m := NewManager(t.TempDir())
	p, err := m.StartEnrollment("admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.VerifyEnrollment("admin", totpCode(p.Secret, time.Now().Unix()/30)); err != nil {
		t.Fatal(err)
	}
	return m, p.Secret
}
func retryState(fn func() error) error {
	for i := 0; i < 500; i++ {
		err := fn()
		if !errors.Is(err, ErrStateConflict) {
			return err
		}
		time.Sleep(time.Millisecond * 2)
	}
	return fmt.Errorf("lock conflict retry exhausted")
}
func loginState(t *testing.T, m *Manager, secret string) string {
	t.Helper()
	token, _, err := m.VerifyLogin("", totpCode(secret, time.Now().Unix()/30))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestAuthConcurrentLoginRevokeAndReaders(t *testing.T) {
	m, secret := enrolledManager(t)
	revoked := loginState(t, m, secret)
	retained := loginState(t, m, secret)
	done := make(chan struct{})
	readErrors := make(chan error, 1)
	var readers sync.WaitGroup
	readers.Add(1)
	go func() {
		defer readers.Done()
		for {
			select {
			case <-done:
				return
			default:
			}
			_, err := NewManager(filepath.Dir(filepath.Dir(m.path))).load()
			if err != nil && !errors.Is(err, ErrStateConflict) {
				select {
				case readErrors <- err:
				default:
				}
				return
			}
		}
	}()
	const count = 24
	tokens := make([]string, count)
	errs := make(chan error, count+1)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range count {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			other := NewManager(filepath.Dir(filepath.Dir(m.path)))
			errs <- retryState(func() error {
				var err error
				tokens[i], _, err = other.VerifyLogin("", totpCode(secret, time.Now().Unix()/30))
				return err
			})
		}(i)
	}
	wg.Add(1)
	go func() { defer wg.Done(); <-start; errs <- retryState(func() error { return m.RevokeToken(revoked) }) }()
	close(start)
	wg.Wait()
	close(done)
	readers.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	select {
	case err := <-readErrors:
		t.Fatal(err)
	default:
	}
	fresh := NewManager(filepath.Dir(filepath.Dir(m.path)))
	if fresh.ValidateToken(revoked) {
		t.Fatal("revoked token resurrected")
	}
	if !fresh.ValidateToken(retained) {
		t.Fatal("other device lost")
	}
	for _, token := range tokens {
		if token == "" || !fresh.ValidateToken(token) {
			t.Fatal("successful login lost")
		}
	}
	state, err := fresh.load()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Sessions) != count+1 {
		t.Fatalf("sessions=%d", len(state.Sessions))
	}
}

func TestAuthStaleEnrollmentCannotReplaceAccount(t *testing.T) {
	dir := t.TempDir()
	a, b := NewManager(dir), NewManager(dir)
	pa, err := a.StartEnrollment("alice")
	if err != nil {
		t.Fatal(err)
	}
	pb, err := b.StartEnrollment("bob")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.VerifyEnrollment("alice", totpCode(pa.Secret, time.Now().Unix()/30)); err != nil {
		t.Fatal(err)
	}
	token := loginState(t, a, pa.Secret)
	before, _ := os.ReadFile(a.path)
	if _, err = b.VerifyEnrollment("bob", totpCode(pb.Secret, time.Now().Unix()/30)); err == nil || err.Error() != "enrollment is already complete" {
		t.Fatalf("stale enrolment: %v", err)
	}
	after, _ := os.ReadFile(a.path)
	if string(before) != string(after) || !b.ValidateToken(token) {
		t.Fatal("stale enrollment changed account")
	}
}

func TestAuthStatePreservesExtensionsAndLastRevoke(t *testing.T) {
	m, secret := enrolledManager(t)
	data, err := os.ReadFile(m.path)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err = json.Unmarshal(data, &object); err != nil {
		t.Fatal(err)
	}
	object["future"] = map[string]any{"revision": 7, "devices": []string{"one", "two"}}
	data, _ = json.Marshal(object)
	if err = os.WriteFile(m.path, data, 0600); err != nil {
		t.Fatal(err)
	}
	token := loginState(t, m, secret)
	if err = m.RevokeToken(token); err != nil {
		t.Fatal(err)
	}
	if err = m.RevokeToken(token); err != nil {
		t.Fatal(err)
	}
	fresh := NewManager(filepath.Dir(filepath.Dir(m.path)))
	state, err := fresh.load()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Sessions) != 0 || fresh.ValidateToken(token) {
		t.Fatal("last revoked token reintroduced")
	}
	if string(state.extra["future"]) != "" {
		var extra map[string]any
		_ = json.Unmarshal(state.extra["future"], &extra)
		if extra["revision"] != float64(7) {
			t.Fatal("unknown field changed")
		}
	} else {
		t.Fatal("unknown field lost")
	}
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(m.path)
		if info.Mode().Perm() != 0600 {
			t.Fatalf("permissions=%v", info.Mode())
		}
	}
}

func TestAuthStateCorruptionAndFailurePreserveBytes(t *testing.T) {
	for _, body := range []string{"{broken", "null", "[]", strings.Repeat("x", maxStateBytes+1)} {
		t.Run(fmt.Sprint(len(body)), func(t *testing.T) {
			m := NewManager(t.TempDir())
			if err := os.MkdirAll(filepath.Dir(m.path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(m.path, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			if _, _, err := m.VerifyLogin("", "123456"); err == nil {
				t.Fatal("corruption accepted")
			}
			if err := m.RevokeToken("opaque"); err == nil {
				t.Fatal("corrupt revoke accepted")
			}
			got, _ := os.ReadFile(m.path)
			if string(got) != body {
				t.Fatal("corrupt bytes rewritten")
			}
		})
	}
	m, _ := enrolledManager(t)
	before, _ := os.ReadFile(m.path)
	if err := m.updateState(func(s *State, _ bool) error { s.Username = "other"; return errors.New("stop") }); err == nil {
		t.Fatal("mutation error swallowed")
	}
	if err := m.updateState(func(s *State, _ bool) error { s.Username = strings.Repeat("x", maxStateBytes); return nil }); err == nil {
		t.Fatal("oversize encoded")
	}
	after, _ := os.ReadFile(m.path)
	if string(before) != string(after) {
		t.Fatal("failed write modified state")
	}
	temps, _ := filepath.Glob(filepath.Join(filepath.Dir(m.path), ".auth-*.tmp"))
	if len(temps) != 0 {
		t.Fatal("temp files leaked")
	}
	// Stale temporary files are not state and do not prevent a later commit.
	if err := os.WriteFile(filepath.Join(filepath.Dir(m.path), ".auth-abandoned.tmp"), []byte("partial"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.RevokeToken("absent"); err != nil {
		t.Fatal(err)
	}
}

func TestAuthStateRejectsUnsafePathsAndOutOfBandChange(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink setup requires Windows privilege; locking/replacement tested separately")
	}
	for _, target := range []string{"directory", "state", "lock"} {
		t.Run(target, func(t *testing.T) {
			dir := t.TempDir()
			m := NewManager(dir)
			outside := t.TempDir()
			data := []byte(`{"username":"outside"}`)
			if err := os.WriteFile(filepath.Join(outside, "state"), data, 0600); err != nil {
				t.Fatal(err)
			}
			if target == "directory" {
				if err := os.Symlink(outside, filepath.Join(dir, ".gi")); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.MkdirAll(filepath.Dir(m.path), 0700); err != nil {
					t.Fatal(err)
				}
				name := "auth.json"
				if target == "lock" {
					name = "auth.lock"
				}
				if err := os.Symlink(filepath.Join(outside, "state"), filepath.Join(dir, ".gi", name)); err != nil {
					t.Fatal(err)
				}
			}
			if err := m.RevokeToken("opaque"); err == nil {
				t.Fatal("unsafe path accepted")
			}
			got, _ := os.ReadFile(filepath.Join(outside, "state"))
			if string(got) != string(data) {
				t.Fatal("outside file altered")
			}
		})
	}
	m, _ := enrolledManager(t)
	external := []byte(`{"username":"external","totp_enabled":false}`)
	err := m.updateState(func(s *State, _ bool) error { s.Username = "lost"; return os.WriteFile(m.path, external, 0600) })
	if !errors.Is(err, ErrStateConflict) {
		t.Fatalf("external writer=%v", err)
	}
	got, _ := os.ReadFile(m.path)
	if string(got) != string(external) {
		t.Fatal("external change overwritten")
	}
}

func TestAuthStateProcessHelper(t *testing.T) {
	dir := os.Getenv("GI_AUTH_TEST_PROCESS")
	if dir == "" {
		return
	}
	m := NewManager(dir)
	if os.Getenv("GI_AUTH_TEST_HOLD") == "1" {
		if err := m.updateState(func(*State, bool) error {
			if err := os.WriteFile(filepath.Join(dir, "held"), []byte("held"), 0600); err != nil {
				return err
			}
			select {}
		}); err != nil {
			t.Fatal(err)
		}
		return
	}
	var token string
	if err := retryState(func() error {
		state, err := m.load()
		if err != nil {
			return err
		}
		token, _, err = m.VerifyLogin("", totpCode(state.TOTPSecret, time.Now().Unix()/30))
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, os.Getenv("GI_AUTH_TEST_OUTPUT")), []byte(token), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestAuthStateCrossProcessAndCrashRelease(t *testing.T) {
	m, _ := enrolledManager(t)
	dir := filepath.Dir(filepath.Dir(m.path))
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	holder := exec.Command(binary, "-test.run=^TestAuthStateProcessHelper$")
	holder.Env = append(os.Environ(), "GI_AUTH_TEST_PROCESS="+dir, "GI_AUTH_TEST_HOLD=1")
	if err = holder.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if holder.ProcessState == nil {
			_ = holder.Process.Kill()
			_ = holder.Wait()
		}
	}()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err = os.Stat(filepath.Join(dir, "held")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("child did not acquire lock")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err = m.RevokeToken("none"); !errors.Is(err, ErrStateConflict) {
		t.Fatalf("process lock not exclusive: %v", err)
	}
	_ = holder.Process.Kill()
	_ = holder.Wait()
	if err = m.RevokeToken("none"); err != nil {
		t.Fatalf("crashed lock retained: %v", err)
	}
	const count = 4
	var wg sync.WaitGroup
	errs := make(chan error, count)
	for i := range count {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cmd := exec.Command(binary, "-test.run=^TestAuthStateProcessHelper$")
			cmd.Env = append(os.Environ(), "GI_AUTH_TEST_PROCESS="+dir, fmt.Sprintf("GI_AUTH_TEST_OUTPUT=token-%d", i))
			output, err := cmd.CombinedOutput()
			if err != nil {
				err = fmt.Errorf("child %d: %w: %s", i, err, output)
			}
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := range count {
		token, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf("token-%d", i)))
		if err != nil {
			t.Fatal(err)
		}
		if !m.ValidateToken(string(token)) {
			t.Fatal("child login lost")
		}
	}
}

func TestAuthStatePinnedDirectoryAndLockCannotBeSwapped(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows denies replacement of open lock/root handles; native lock tests run in CI")
	}
	for _, name := range []string{"directory", "lock"} {
		t.Run(name, func(t *testing.T) {
			m, _ := enrolledManager(t)
			before, err := os.ReadFile(m.path)
			if err != nil {
				t.Fatal(err)
			}
			dir := filepath.Dir(m.path)
			old := dir + "-old"
			err = m.updateState(func(state *State, _ bool) error {
				state.Username = "must-not-commit"
				if name == "lock" {
					if err := os.Rename(filepath.Join(dir, "auth.lock"), filepath.Join(dir, "auth-old.lock")); err != nil {
						return err
					}
					return os.WriteFile(filepath.Join(dir, "auth.lock"), nil, 0600)
				}
				if err := os.Rename(dir, old); err != nil {
					return err
				}
				return os.Mkdir(dir, 0700)
			})
			if !errors.Is(err, ErrStateConflict) {
				t.Fatalf("replacement=%v", err)
			}
			path := m.path
			if name == "directory" {
				path = filepath.Join(old, "auth.json")
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatal("pinned original modified")
			}
		})
	}
}

func TestAuthStateConcurrentEnrollmentHasOneWinner(t *testing.T) {
	dir := t.TempDir()
	managers := []*Manager{NewManager(dir), NewManager(dir)}
	pending := make([]PendingEnrollment, 2)
	for i, m := range managers {
		var err error
		pending[i], err = m.StartEnrollment(fmt.Sprintf("user%d", i))
		if err != nil {
			t.Fatal(err)
		}
	}
	errs := make(chan error, 2)
	start := make(chan struct{})
	for i, m := range managers {
		go func(i int, m *Manager) {
			<-start
			errs <- retryState(func() error {
				_, err := m.VerifyEnrollment(pending[i].Username, totpCode(pending[i].Secret, time.Now().Unix()/30))
				return err
			})
		}(i, m)
	}
	close(start)
	wins := 0
	for range managers {
		err := <-errs
		if err == nil {
			wins++
		} else if err.Error() != "enrollment is already complete" {
			t.Fatal(err)
		}
	}
	if wins != 1 {
		t.Fatalf("winners=%d", wins)
	}
	state, err := NewManager(dir).load()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pending {
		if p.Username == state.Username && p.Secret != state.TOTPSecret {
			t.Fatal("mixed enrollment")
		}
	}
}

// Used by package fixtures; production writes use updateState with a mutation.
func (m *Manager) save(state State) error {
	return m.updateState(func(current *State, _ bool) error { state.extra = current.extra; *current = state; return nil })
}
