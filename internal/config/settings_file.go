package config

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

var settingsMu sync.Mutex
var ErrSettingsConflict = errors.New("Pi settings changed or are locked; reload saved policy and retry")

const settingsMaxBytes = 1 << 20

type settingsDocument struct {
	Values   map[string]json.RawMessage
	Revision string
	Mode     os.FileMode
}

func readSettingsDocument(root *os.Root) (settingsDocument, error) {
	d := settingsDocument{Values: map[string]json.RawMessage{}, Revision: "missing", Mode: 0600}
	info, err := root.Lstat("settings.json")
	if errors.Is(err, os.ErrNotExist) {
		return d, nil
	}
	if err != nil {
		return d, err
	}
	if !info.Mode().IsRegular() || info.Size() > settingsMaxBytes {
		return d, errors.New("settings.json must be a regular non-symlink file no larger than 1 MiB")
	}
	f, err := root.Open("settings.json")
	if err != nil {
		return d, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil {
		return d, err
	}
	if !os.SameFile(info, opened) {
		return d, ErrSettingsConflict
	}
	raw, err := io.ReadAll(io.LimitReader(f, settingsMaxBytes+1))
	if err != nil {
		return d, err
	}
	if len(raw) > settingsMaxBytes {
		return d, errors.New("settings.json exceeds 1 MiB")
	}
	if err = json.Unmarshal(raw, &d.Values); err != nil || d.Values == nil {
		return d, errors.New("settings.json must contain a valid JSON object")
	}
	hash := sha256.Sum256(raw)
	d.Revision = hex.EncodeToString(hash[:])
	d.Mode = info.Mode().Perm()
	return d, nil
}

func readPiSettings(workspace string) (settingsDocument, error) {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	root, err := openConfigDirectory(workspace, ".pi", false)
	if errors.Is(err, os.ErrNotExist) {
		return settingsDocument{Values: map[string]json.RawMessage{}, Revision: "missing", Mode: 0600}, nil
	}
	if err != nil {
		return settingsDocument{}, err
	}
	defer root.Close()
	return readSettingsDocument(root)
}

// All native .pi/settings.json mutations cooperate on this lock. expected empty
// means an atomic merge of current values; policy writes require a file revision.
func updatePiSettings(workspace, expected string, change func(*settingsDocument) error) (settingsDocument, error) {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	root, err := openConfigDirectory(workspace, ".pi", true)
	if err != nil {
		return settingsDocument{}, err
	}
	defer root.Close()
	if info, e := root.Lstat(".gi-settings.lock"); e == nil && !info.Mode().IsRegular() {
		return settingsDocument{}, errors.New("settings lock must be regular")
	} else if e != nil && !errors.Is(e, os.ErrNotExist) {
		return settingsDocument{}, e
	}
	lock, err := root.OpenFile(".gi-settings.lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return settingsDocument{}, err
	}
	defer lock.Close()
	opened, err := lock.Stat()
	current, e := root.Lstat(".gi-settings.lock")
	if err != nil || e != nil || !current.Mode().IsRegular() || !os.SameFile(opened, current) {
		return settingsDocument{}, ErrSettingsConflict
	}
	if err = lockIdentityFile(lock); err != nil {
		if errors.Is(err, ErrIdentityConflict) {
			err = ErrSettingsConflict
		}
		return settingsDocument{}, err
	}
	defer unlockIdentityFile(lock)
	d, err := readSettingsDocument(root)
	if err != nil {
		return d, err
	}
	if expected != "" && d.Revision != expected {
		return d, ErrSettingsConflict
	}
	originalRevision := d.Revision
	if err = change(&d); err != nil {
		return d, err
	}
	body, err := json.MarshalIndent(d.Values, "", "  ")
	if err != nil {
		return d, err
	}
	body = append(body, '\n')
	if len(body) > settingsMaxBytes {
		return d, errors.New("updated settings exceed 1 MiB")
	}
	var nonce [12]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return d, err
	}
	temp := ".gi-settings-" + hex.EncodeToString(nonce[:]) + ".tmp"
	file, err := root.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return d, err
	}
	defer root.Remove(temp)
	if err = file.Chmod(d.Mode); err == nil {
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
		return d, err
	}
	latest, err := readSettingsDocument(root)
	if err != nil {
		return d, err
	}
	if latest.Revision != originalRevision {
		return d, ErrSettingsConflict
	}
	if err = root.Rename(temp, "settings.json"); err != nil {
		return d, err
	}
	dir, err := root.Open(".")
	if err == nil {
		err = dir.Sync()
		dir.Close()
	}
	if err != nil {
		return d, fmt.Errorf("settings replaced but directory sync failed; reload to verify: %w", err)
	}
	hash := sha256.Sum256(body)
	d.Revision = hex.EncodeToString(hash[:])
	return d, nil
}

func persistPiFields(workspace string, fields map[string]any) error {
	_, err := updatePiSettings(workspace, "", func(d *settingsDocument) error {
		for key, value := range fields {
			raw, err := json.Marshal(value)
			if err != nil {
				return err
			}
			d.Values[key] = raw
		}
		return nil
	})
	return err
}
