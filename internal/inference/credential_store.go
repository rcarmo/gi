package inference

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
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

func openCredentialRoot(create bool) (*os.Root, error) {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return nil, errors.New("credential home unavailable")
	}
	root, err := os.OpenRoot(home)
	if err != nil {
		return nil, errors.New("credential home unavailable")
	}
	for _, name := range []string{".pi", "agent"} {
		info, e := root.Lstat(name)
		if errors.Is(e, os.ErrNotExist) && create {
			e = root.Mkdir(name, 0700)
			if e == nil || errors.Is(e, os.ErrExist) {
				info, e = root.Lstat(name)
			}
		}
		if e != nil {
			root.Close()
			return nil, e
		}
		if !info.IsDir() {
			root.Close()
			return nil, errors.New("credential directory must not be a symlink")
		}
		next, e := root.OpenRoot(name)
		if e != nil {
			root.Close()
			return nil, errors.New("cannot open credential directory")
		}
		opened, e := next.Stat(".")
		current, currentErr := root.Lstat(name)
		root.Close()
		if e != nil || currentErr != nil || !current.IsDir() || !os.SameFile(info, opened) || !os.SameFile(opened, current) {
			next.Close()
			return nil, ErrCredentialConflict
		}
		root = next
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

// Both web API-key changes and native TUI logout cooperate on this writer.
// No external provider call or credential verification is performed here.
func updateCredentials(expected string, change func(map[string]json.RawMessage) error) (credentialDocument, error) {
	credentialMu.Lock()
	defer credentialMu.Unlock()
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
	if err = change(d.entries); err != nil {
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
		entry["apiKey"], _ = json.Marshal(key)
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
