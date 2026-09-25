package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var ErrStateConflict = errors.New("authentication state changed; retry")

const maxStateBytes = 1 << 20

// Pin .gi under the configured workspace and reject symlink state directories.
// The caller owns the returned roots; roots also keep path operations bounded.
func (m *Manager) openStateRoot(create bool) (*os.Root, *os.Root, error) {
	workspace, err := os.OpenRoot(filepath.Dir(filepath.Dir(m.path)))
	if err != nil {
		return nil, nil, err
	}
	fail := func(err error) (*os.Root, *os.Root, error) { workspace.Close(); return nil, nil, err }
	info, err := workspace.Lstat(".gi")
	if os.IsNotExist(err) && create {
		if err = workspace.Mkdir(".gi", 0700); err != nil && !os.IsExist(err) {
			return fail(err)
		}
		info, err = workspace.Lstat(".gi")
	}
	if err != nil {
		return fail(err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fail(errors.New("auth directory must be a real directory"))
	}
	root, err := workspace.OpenRoot(".gi")
	if err != nil {
		return fail(err)
	}
	opened, err := root.Stat(".")
	if err != nil || !os.SameFile(info, opened) {
		root.Close()
		return fail(ErrStateConflict)
	}
	// Persist a newly created .gi entry as well as the auth file inside it.
	// Sync even when it already exists: another process may have just created it.
	if create {
		if err := syncStateDir(workspace); err != nil {
			root.Close()
			return fail(err)
		}
	}
	return workspace, root, nil
}

func readState(root *os.Root) (State, []byte, error) {
	var state State
	info, err := root.Lstat("auth.json")
	if err != nil {
		return state, nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxStateBytes {
		return state, nil, errors.New("auth state must be a bounded regular file")
	}
	file, err := root.Open("auth.json")
	if err != nil {
		return state, nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return state, nil, err
	}
	if !os.SameFile(info, opened) {
		return state, nil, ErrStateConflict
	}
	data, err := io.ReadAll(io.LimitReader(file, maxStateBytes+1))
	if err != nil {
		return state, nil, err
	}
	if len(data) > maxStateBytes {
		return state, nil, errors.New("auth state is too large")
	}
	var object map[string]json.RawMessage
	if err = json.Unmarshal(data, &object); err != nil {
		return state, nil, err
	}
	if object == nil {
		return state, nil, errors.New("auth state must be an object")
	}
	if err = json.Unmarshal(data, &state); err != nil {
		return state, nil, err
	}
	// Known optional keys must not be reintroduced from the old document (notably
	// an omitted empty sessions list after revoking its final token).
	for _, key := range []string{"username", "totp_secret", "totp_enabled", "created_at", "updated_at", "sessions", "webauthn_user_id", "passkeys", "webauthn_ceremonies", "login_policy"} {
		delete(object, key)
	}
	state.extra = object
	return state, data, nil
}

func encodeState(state State) ([]byte, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	var object map[string]json.RawMessage
	if err = json.Unmarshal(data, &object); err != nil {
		return nil, err
	}
	for key, value := range state.extra {
		if _, exists := object[key]; !exists {
			object[key] = value
		}
	}
	data, err = json.MarshalIndent(object, "", "  ")
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	if len(data) > maxStateBytes {
		return nil, errors.New("auth state is too large")
	}
	return data, nil
}

// Separate first creation from opening an existing lock. Concurrent initial
// creators must converge on one inode; never truncate or replace the lock.
func openStateLock(root *os.Root) (*os.File, error) {
	lock, err := root.OpenFile("auth.lock", os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if os.IsExist(err) {
		lock, err = root.OpenFile("auth.lock", os.O_RDWR, 0)
	}
	// A disappearing pathname (including a concurrent first-open race on
	// Darwin) is retryable. Each retry reopens and revalidates the rooted path.
	if os.IsNotExist(err) {
		return nil, ErrStateConflict
	}
	return lock, err
}

// All production mutators must use this transaction. Lockless readers see one
// complete snapshot. A validation already in flight may see the previous token
// set; a fresh validation after a committed revoke observes its removal.
func (m *Manager) updateState(change func(*State, bool) error) error {
	workspace, root, err := m.openStateRoot(true)
	if err != nil {
		return err
	}
	defer workspace.Close()
	defer root.Close()
	lockInfo, err := root.Lstat("auth.lock")
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil && !lockInfo.Mode().IsRegular() {
		return errors.New("auth lock must be a regular file")
	}
	lock, err := openStateLock(root)
	if err != nil {
		return err
	}
	defer lock.Close()
	opened, err := lock.Stat()
	if err != nil {
		return err
	}
	lockInfo, err = root.Lstat("auth.lock")
	if err != nil || !lockInfo.Mode().IsRegular() || !os.SameFile(opened, lockInfo) {
		return ErrStateConflict
	}
	if err = lockStateFile(lock); err != nil {
		return err
	}
	defer unlockStateFile(lock)
	state, before, err := readState(root)
	exists := err == nil
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err = change(&state, exists); err != nil {
		return err
	}
	after, err := encodeState(state)
	if err != nil {
		return err
	}
	nonce, _, err := newToken()
	if err != nil {
		return err
	}
	name := ".auth-" + nonce + ".tmp"
	file, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer root.Remove(name)
	if _, err = file.Write(after); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	// Recheck pathname ownership and reject a non-cooperating writer before the
	// commit. Local filesystem owners are not an adversarial security boundary.
	currentDir, err := workspace.Lstat(".gi")
	if err != nil {
		return err
	}
	pinnedDir, err := root.Stat(".")
	if err != nil {
		return err
	}
	if !currentDir.IsDir() || !os.SameFile(currentDir, pinnedDir) {
		return ErrStateConflict
	}
	lockInfo, err = root.Lstat("auth.lock")
	if err != nil || !lockInfo.Mode().IsRegular() || !os.SameFile(opened, lockInfo) {
		return ErrStateConflict
	}
	_, current, readErr := readState(root)
	if exists {
		if readErr != nil || !bytes.Equal(before, current) {
			return ErrStateConflict
		}
	} else if !os.IsNotExist(readErr) {
		return ErrStateConflict
	}
	if err = root.Rename(name, "auth.json"); err != nil {
		return fmt.Errorf("replace auth state: %w", err)
	}
	if err = syncStateDir(root); err != nil {
		return fmt.Errorf("auth state replaced but durability unconfirmed: %w", err)
	}
	return nil
}

func (m *Manager) load() (State, error) {
	workspace, root, err := m.openStateRoot(false)
	if err != nil {
		return State{}, err
	}
	defer workspace.Close()
	defer root.Close()
	state, _, err := readState(root)
	return state, err
}
