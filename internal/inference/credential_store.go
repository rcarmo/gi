package inference

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/rcarmo/gi/internal/lockdir"
)

var credentialMu sync.Mutex
var credentialRevisionOnce sync.Once
var credentialRevisionKey [32]byte
var credentialRevisionErr error
var ErrCredentialConflict = errors.New("credential store changed or is busy; refresh providers and retry")
var ErrCredentialInvalid = errors.New("invalid or unsupported API-key change")

const credentialMaxBytes = 1 << 20

type credentialDocument struct {
	entries  map[string]json.RawMessage
	revision string
}

func credentialRevision(raw []byte) (string, error) {
	credentialRevisionOnce.Do(func() { _, credentialRevisionErr = rand.Read(credentialRevisionKey[:]) })
	if credentialRevisionErr != nil {
		return "", errors.New("credential revision unavailable")
	}
	mac := hmac.New(sha256.New, credentialRevisionKey[:])
	mac.Write(raw)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// openCredentialRoot opens the directory of the credentials file
// (AuthFilePath: ~/.gi/agent when it has auth.json, else Pi's agent
// directory), creating Pi's when asked. The directory itself must not be a
// symlink, and must still be the one opened.
func openCredentialRoot(create bool) (*os.Root, error) {
	dir := filepath.Dir(AuthFilePath())
	if strings.TrimSpace(dir) == "" || dir == "." {
		return nil, errors.New("credential home unavailable")
	}
	info, err := os.Lstat(dir)
	if errors.Is(err, os.ErrNotExist) && create {
		if err = os.MkdirAll(dir, 0700); err == nil {
			info, err = os.Lstat(dir)
		}
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("credential directory must not be a symlink")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, errors.New("cannot open credential directory")
	}
	opened, err := root.Stat(".")
	if current, e := os.Lstat(dir); err != nil || e != nil || !current.IsDir() || !os.SameFile(info, opened) || !os.SameFile(opened, current) {
		root.Close()
		return nil, ErrCredentialConflict
	}
	return root, nil
}

func readCredentialDocument(root *os.Root) (credentialDocument, error) {
	d := credentialDocument{entries: map[string]json.RawMessage{}}
	info, err := root.Lstat("auth.json")
	if errors.Is(err, os.ErrNotExist) {
		d.revision, err = credentialRevision([]byte("missing"))
		return d, err
	}
	if err != nil {
		return d, errors.New("cannot inspect credential file")
	}
	if !info.Mode().IsRegular() || info.Size() > credentialMaxBytes {
		return d, errors.New("credential file must be a regular non-symlink file no larger than 1 MiB")
	}
	f, err := root.Open("auth.json")
	if err != nil {
		return d, errors.New("cannot open credential file")
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return d, ErrCredentialConflict
	}
	raw, err := io.ReadAll(io.LimitReader(f, credentialMaxBytes+1))
	if err != nil || len(raw) > credentialMaxBytes {
		return d, errors.New("cannot read bounded credential file")
	}
	if err = json.Unmarshal(raw, &d.entries); err != nil || d.entries == nil {
		return d, errors.New("credential file must be a JSON object")
	}
	for _, raw := range d.entries {
		var object map[string]json.RawMessage
		if err = json.Unmarshal(raw, &object); err != nil || object == nil {
			return d, errors.New("credential entries must be JSON objects")
		}
	}
	d.revision, err = credentialRevision(raw)
	return d, err
}

func readCredentials() (credentialDocument, error) {
	credentialMu.Lock()
	defer credentialMu.Unlock()
	root, err := openCredentialRoot(false)
	if errors.Is(err, os.ErrNotExist) {
		revision, e := credentialRevision([]byte("missing"))
		return credentialDocument{entries: map[string]json.RawMessage{}, revision: revision}, e
	}
	if err != nil {
		return credentialDocument{}, err
	}
	defer root.Close()
	return readCredentialDocument(root)
}

// errCredentialUnchanged, returned by an updateCredentials change, leaves
// the file as it is.
var errCredentialUnchanged = errors.New("credentials unchanged")

// piAuthLockStale is Pi's FileAuthStorageBackend lock staleness and wait.
const piAuthLockStale = 30 * time.Second

// Web API-key changes, native TUI logout and OAuth refreshes cooperate on
// this writer, which also holds Pi's lock (auth.json.lock) so gi and Pi do
// not interleave writes. A change returning errCredentialUnchanged skips the
// write.
func updateCredentials(expected string, change func(map[string]json.RawMessage) error) (credentialDocument, error) {
	return updateCredentialsContext(context.Background(), expected, change)
}

// Cancellable acquisition matters for per-request provider refresh: another
// process may hold Pi's lock, or another goroutine may be refreshing a token.
func lockCredentialsContext(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if credentialMu.TryLock() {
			return nil
		}
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func updateCredentialsContext(ctx context.Context, expected string, change func(map[string]json.RawMessage) error) (d credentialDocument, err error) {
	if err := lockCredentialsContext(ctx); err != nil {
		return d, err
	}
	defer credentialMu.Unlock()
	lockErr := lockdir.WithContext(ctx, filepath.Join(filepath.Dir(AuthFilePath()), "auth.json.lock"), piAuthLockStale, piAuthLockStale, func() error {
		if err = ctx.Err(); err == nil {
			d, err = updateCredentialsLocked(expected, change)
		}
		return nil
	})
	if lockErr != nil {
		if ctx.Err() != nil {
			return d, ctx.Err()
		}
		return d, ErrCredentialConflict
	}
	return d, err
}

func updateCredentialsLocked(expected string, change func(map[string]json.RawMessage) error) (credentialDocument, error) {
	root, err := openCredentialRoot(true)
	if err != nil {
		return credentialDocument{}, err
	}
	defer root.Close()
	if info, e := root.Lstat(".gi-auth.lock"); e == nil && !info.Mode().IsRegular() {
		return credentialDocument{}, errors.New("credential lock must be regular")
	} else if e != nil && !errors.Is(e, os.ErrNotExist) {
		return credentialDocument{}, errors.New("cannot inspect credential lock")
	}
	lock, err := root.OpenFile(".gi-auth.lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return credentialDocument{}, errors.New("cannot open credential lock")
	}
	defer lock.Close()
	opened, err := lock.Stat()
	current, e := root.Lstat(".gi-auth.lock")
	if err != nil || e != nil || !current.Mode().IsRegular() || !os.SameFile(opened, current) {
		return credentialDocument{}, ErrCredentialConflict
	}
	if err = lockCredentialFile(lock); err != nil {
		return credentialDocument{}, err
	}
	defer unlockCredentialFile(lock)
	d, err := readCredentialDocument(root)
	if err != nil {
		return d, err
	}
	if expected != "" && expected != d.revision {
		return d, ErrCredentialConflict
	}
	revision := d.revision
	if err = change(d.entries); errors.Is(err, errCredentialUnchanged) {
		return d, nil
	} else if err != nil {
		return d, err
	}
	body, err := json.MarshalIndent(d.entries, "", "  ")
	if err != nil {
		return d, errors.New("cannot encode credentials")
	}
	body = append(body, '\n')
	if len(body) > credentialMaxBytes {
		return d, errors.New("credential file exceeds 1 MiB")
	}
	var nonce [12]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return d, errors.New("cannot prepare credential write")
	}
	temp := ".gi-auth-" + hex.EncodeToString(nonce[:]) + ".tmp"
	file, err := root.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return d, errors.New("cannot prepare credential file")
	}
	defer root.Remove(temp)
	if err = file.Chmod(0600); err == nil {
		_, err = file.Write(body)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return d, errors.New("cannot persist credentials")
	}
	latest, err := readCredentialDocument(root)
	if err != nil {
		return d, err
	}
	if latest.revision != revision {
		return d, ErrCredentialConflict
	}
	if err = root.Rename(temp, "auth.json"); err != nil {
		return d, errors.New("cannot replace credential file")
	}
	dir, err := root.Open(".")
	if err == nil {
		err = dir.Sync()
		dir.Close()
	}
	if err != nil {
		return d, errors.New("credentials replaced but directory sync failed; refresh to verify")
	}
	d.revision, err = credentialRevision(body)
	return d, err
}

func APIKeyProvider(provider string) bool { return provider == "openai" || provider == "anthropic" }
func editableKeyEntry(raw json.RawMessage) bool {
	if raw == nil {
		return true
	}
	var entry authEntry
	if json.Unmarshal(raw, &entry) != nil {
		return false
	}
	return (entry.Type == "api_key" || entry.Type == "api-key") && entry.Access == "" && entry.Refresh == "" && entry.Token == ""
}
func SaveProviderKey(provider, revision, key string) error {
	if !APIKeyProvider(provider) || revision == "" || len(key) < 1 || len(key) > 4096 || !utf8.ValidString(key) {
		return ErrCredentialInvalid
	}
	for _, r := range key {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return ErrCredentialInvalid
		}
	}
	_, err := updateCredentials(revision, func(entries map[string]json.RawMessage) error {
		if !editableKeyEntry(entries[provider]) {
			return ErrCredentialInvalid
		}
		entry := map[string]json.RawMessage{}
		if raw := entries[provider]; raw != nil {
			_ = json.Unmarshal(raw, &entry)
		}
		entry["type"], _ = json.Marshal("api_key")
		entry["key"], _ = json.Marshal(key) // Pi's field
		delete(entry, "apiKey")
		entries[provider], _ = json.Marshal(entry)
		return nil
	})
	return err
}
func RemoveProviderKey(provider, revision string) error {
	if !APIKeyProvider(provider) || revision == "" {
		return ErrCredentialInvalid
	}
	_, err := updateCredentials(revision, func(entries map[string]json.RawMessage) error {
		if !editableKeyEntry(entries[provider]) {
			return ErrCredentialInvalid
		}
		delete(entries, provider)
		return nil
	})
	return err
}
