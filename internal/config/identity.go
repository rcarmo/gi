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
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

const identityMaxBytes = 1 << 20

var identityMu sync.Mutex
var ErrIdentityConflict = errors.New("saved configuration changed; reload saved names before retrying")
var ErrIdentityInvalid = errors.New("invalid display names")

type IdentityNames struct {
	AssistantName string `json:"assistant_name"`
	UserName      string `json:"user_name"`
}
type IdentitySnapshot struct {
	IdentityNames
	Revision string `json:"revision"`
}

type identityDocument struct {
	values          map[string]json.RawMessage
	assistant, user map[string]json.RawMessage
	snapshot        IdentitySnapshot
	mode            os.FileMode
}

func validateIdentity(names IdentityNames) (IdentityNames, error) {
	for _, name := range []string{names.AssistantName, names.UserName} {
		if !utf8.ValidString(name) || strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 128 {
			return names, fmt.Errorf("%w: each name must contain 1–128 characters", ErrIdentityInvalid)
		}
		for _, r := range name {
			if unicode.IsControl(r) {
				return names, fmt.Errorf("%w: control characters are not allowed", ErrIdentityInvalid)
			}
		}
	}
	names.AssistantName = strings.TrimSpace(names.AssistantName)
	names.UserName = strings.TrimSpace(names.UserName)
	return names, nil
}

func openIdentityRoot(workspace string, create bool) (*os.Root, error) {
	if strings.TrimSpace(workspace) == "" {
		return nil, errors.New("workspace root is required")
	}
	root, err := os.OpenRoot(workspace)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	info, err := root.Lstat(".piclaw")
	if errors.Is(err, os.ErrNotExist) && create {
		if err = root.Mkdir(".piclaw", 0700); err != nil && !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		info, err = root.Lstat(".piclaw")
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New(".piclaw must be a real directory")
	}
	configRoot, err := root.OpenRoot(".piclaw")
	if err != nil {
		return nil, err
	}
	opened, err := configRoot.Stat(".")
	current, currentErr := root.Lstat(".piclaw")
	if err != nil || currentErr != nil || !current.IsDir() || !os.SameFile(info, opened) || !os.SameFile(opened, current) {
		configRoot.Close()
		return nil, ErrIdentityConflict
	}
	return configRoot, nil
}

func readIdentityDocument(root *os.Root) (identityDocument, error) {
	d := identityDocument{values: map[string]json.RawMessage{}, assistant: map[string]json.RawMessage{}, user: map[string]json.RawMessage{}, mode: 0600}
	d.snapshot = IdentitySnapshot{IdentityNames: IdentityNames{"Gi", "User"}, Revision: "missing"}
	info, err := root.Lstat("config.json")
	if errors.Is(err, os.ErrNotExist) {
		return d, nil
	}
	if err != nil {
		return d, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > identityMaxBytes {
		return d, errors.New("config.json must be a regular non-symlink file no larger than 1 MiB")
	}
	f, err := root.Open("config.json")
	if err != nil {
		return d, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil {
		return d, err
	}
	if !os.SameFile(info, opened) {
		return d, ErrIdentityConflict
	}
	raw, err := io.ReadAll(io.LimitReader(f, identityMaxBytes+1))
	if err != nil {
		return d, err
	}
	if len(raw) > identityMaxBytes {
		return d, errors.New("config.json exceeds 1 MiB")
	}
	if err = json.Unmarshal(raw, &d.values); err != nil || d.values == nil {
		return d, errors.New("config.json must contain a valid JSON object")
	}
	for key, target := range map[string]*map[string]json.RawMessage{"assistant": &d.assistant, "user": &d.user} {
		if b, ok := d.values[key]; ok {
			if err = json.Unmarshal(b, target); err != nil || *target == nil {
				return d, fmt.Errorf("config %s must be an object", key)
			}
		}
	}
	for key, target := range map[string]*string{"assistantName": &d.snapshot.AssistantName, "userName": &d.snapshot.UserName} {
		section := d.assistant
		if key == "userName" {
			section = d.user
		}
		if b, ok := section[key]; ok {
			var s string
			if err = json.Unmarshal(b, &s); err != nil {
				return d, errors.New("configured display name must be a string")
			}
			if s != "" {
				*target = s
			}
		}
	}
	digest := sha256.Sum256(raw)
	d.snapshot.Revision = hex.EncodeToString(digest[:])
	d.mode = info.Mode().Perm()
	return d, nil
}

func ReadIdentity(workspace string) (IdentitySnapshot, error) {
	identityMu.Lock()
	defer identityMu.Unlock()
	root, err := openIdentityRoot(workspace, false)
	if errors.Is(err, os.ErrNotExist) {
		return IdentitySnapshot{IdentityNames: IdentityNames{"Gi", "User"}, Revision: "missing"}, nil
	}
	if err != nil {
		return IdentitySnapshot{}, err
	}
	defer root.Close()
	d, err := readIdentityDocument(root)
	return d.snapshot, err
}

// SaveIdentity serialises cooperating writers and detects file revisions. External
// editors that ignore the advisory lock can still race the final check/rename.
func SaveIdentity(workspace, revision string, names IdentityNames) (IdentitySnapshot, error) {
	names, err := validateIdentity(names)
	if err != nil {
		return IdentitySnapshot{}, err
	}
	if revision == "" {
		return IdentitySnapshot{}, ErrIdentityConflict
	}
	identityMu.Lock()
	defer identityMu.Unlock()
	root, err := openIdentityRoot(workspace, true)
	if err != nil {
		return IdentitySnapshot{}, err
	}
	defer root.Close()
	// Reject link-based lock redirection, including links within the config root.
	if info, e := root.Lstat(".gi-identity.lock"); e == nil && !info.Mode().IsRegular() {
		return IdentitySnapshot{}, errors.New("identity lock must be a regular file")
	} else if e != nil && !errors.Is(e, os.ErrNotExist) {
		return IdentitySnapshot{}, e
	}
	lock, err := root.OpenFile(".gi-identity.lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return IdentitySnapshot{}, err
	}
	defer lock.Close()
	openedLock, err := lock.Stat()
	currentLock, currentErr := root.Lstat(".gi-identity.lock")
	if err != nil || currentErr != nil || !currentLock.Mode().IsRegular() || !os.SameFile(openedLock, currentLock) {
		return IdentitySnapshot{}, ErrIdentityConflict
	}
	if err = lockIdentityFile(lock); err != nil {
		return IdentitySnapshot{}, err
	}
	defer unlockIdentityFile(lock)
	d, err := readIdentityDocument(root)
	if err != nil {
		return IdentitySnapshot{}, err
	}
	if d.snapshot.Revision != revision {
		return IdentitySnapshot{}, ErrIdentityConflict
	}
	d.assistant["assistantName"], _ = json.Marshal(names.AssistantName)
	d.user["userName"], _ = json.Marshal(names.UserName)
	d.values["assistant"], _ = json.Marshal(d.assistant)
	d.values["user"], _ = json.Marshal(d.user)
	body, err := json.MarshalIndent(d.values, "", "  ")
	if err != nil {
		return IdentitySnapshot{}, err
	}
	body = append(body, '\n')
	if len(body) > identityMaxBytes {
		return IdentitySnapshot{}, errors.New("updated configuration exceeds 1 MiB")
	}
	var nonce [12]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return IdentitySnapshot{}, err
	}
	temp := ".gi-identity-" + hex.EncodeToString(nonce[:]) + ".tmp"
	file, err := root.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return IdentitySnapshot{}, err
	}
	defer root.Remove(temp)
	if err = file.Chmod(d.mode); err == nil {
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
		return IdentitySnapshot{}, err
	}
	latest, err := readIdentityDocument(root)
	if err != nil {
		return IdentitySnapshot{}, err
	}
	if latest.snapshot.Revision != revision {
		return IdentitySnapshot{}, ErrIdentityConflict
	}
	if err = root.Rename(temp, "config.json"); err != nil {
		return IdentitySnapshot{}, err
	}
	directory, err := root.Open(".")
	if err == nil {
		err = directory.Sync()
		directory.Close()
	}
	if err != nil {
		return IdentitySnapshot{}, fmt.Errorf("config replaced but directory sync failed; reload to verify: %w", err)
	}
	digest := sha256.Sum256(body)
	return IdentitySnapshot{IdentityNames: names, Revision: hex.EncodeToString(digest[:])}, nil
}
