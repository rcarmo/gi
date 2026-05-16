package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestAuthTOTPVerifyReturnsInternalServerErrorForCorruptState(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".gi"), 0o755); err != nil {
		t.Fatalf("create .gi dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gi", "auth.json"), []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("write corrupt auth file: %v", err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: root})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/totp/verify", bytes.NewBufferString(`{"username":"rui","code":"000000"}`))
	res := httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for corrupt auth state, got %d body=%s", res.Code, res.Body.String())
	}
}

func TestAuthEnrollVerifyReturnsInternalServerErrorForPersistenceFailure(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".gi"), 0o755); err != nil {
		t.Fatalf("create .gi dir: %v", err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: root})
	pending, err := srv.auth.StartEnrollment("rui")
	if err != nil {
		t.Fatalf("start enrollment: %v", err)
	}
	code := webTestTOTPCode(pending.Secret)
	if err := os.Mkdir(filepath.Join(root, ".gi", "auth.json"), 0o700); err != nil {
		t.Fatalf("replace auth.json with directory: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/enroll/verify", bytes.NewBufferString(`{"username":"rui","code":"`+code+`"}`))
	req.RemoteAddr = "127.0.0.1:1234"
	res := httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for enrollment persistence failure, got %d body=%s", res.Code, res.Body.String())
	}
}

func webTestTOTPCode(secret string) string {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		key, err = base32.StdEncoding.DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
		if err != nil {
			return ""
		}
	}
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], uint64(time.Now().UTC().Unix()/30))
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(msg[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	bin := (uint32(sum[offset])&0x7f)<<24 | (uint32(sum[offset+1])&0xff)<<16 | (uint32(sum[offset+2])&0xff)<<8 | (uint32(sum[offset+3]) & 0xff)
	otp := bin % 1000000
	return fmt.Sprintf("%06d", otp)
}

func TestAuthEnrollStartReturnsInternalServerErrorForCorruptState(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".gi"), 0o755); err != nil {
		t.Fatalf("create .gi dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gi", "auth.json"), []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("write corrupt auth file: %v", err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: root})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/enroll/start", bytes.NewBufferString(`{"username":"rui"}`))
	req.RemoteAddr = "127.0.0.1:1234"
	res := httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for corrupt auth state, got %d body=%s", res.Code, res.Body.String())
	}
}
