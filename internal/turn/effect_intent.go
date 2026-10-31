package turn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Pin JSON values rather than mutable hook-owned maps. Transport credentials
// and headers never belong in a recoverable provider intent.
func sanitizedProviderIntent(value any) (any, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if len(raw) > 64<<20 {
		return nil, fmt.Errorf("provider intent exceeds 64 MiB")
	}
	// Most prepared contexts contain no credential keys. Keep their already
	// owned JSON rather than decoding every tool/schema/message to mutable maps.
	if !providerIntentHasSecretKey(raw) {
		return json.RawMessage(raw), nil
	}
	var owned any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err = decoder.Decode(&owned); err != nil {
		return nil, err
	}
	var sanitize func(any)
	sanitize = func(value any) {
		switch object := value.(type) {
		case map[string]any:
			for key, child := range object {
				if providerIntentSecretKey(key) {
					delete(object, key)
				} else {
					sanitize(child)
				}
			}
		case []any:
			for _, child := range object {
				sanitize(child)
			}
		}
	}
	sanitize(owned)
	return owned, nil
}

func providerIntentSecretKey(key string) bool {
	switch strings.ToLower(key) {
	case "authorization", "proxy-authorization", "api_key", "apikey", "api-key", "x-api-key", "access_token", "refresh_token", "cookie", "set-cookie":
		return true
	}
	return false
}

// raw has already passed json.Marshal validation. Inspect JSON keys only, not
// words in string values. Escaped keys use the standard decoder; this scanner
// neither interprets nor rewrites JSON values or numerical representations.
func providerIntentHasSecretKey(raw []byte) bool {
	for i := 0; i < len(raw); i++ {
		if raw[i] != '"' {
			continue
		}
		start := i
		escaped := false
		for i++; i < len(raw) && raw[i] != '"'; i++ {
			if raw[i] == '\\' {
				escaped = true
				i++
			}
		}
		end := i
		j := i + 1
		for j < len(raw) && (raw[j] == ' ' || raw[j] == '\n' || raw[j] == '\r' || raw[j] == '\t') {
			j++
		}
		if j >= len(raw) || raw[j] != ':' {
			continue
		}
		if escaped {
			var key string
			if err := json.Unmarshal(raw[start:end+1], &key); err != nil {
				return true
			}
			if providerIntentSecretKey(key) {
				return true
			}
		} else if providerIntentSecretKeyBytes(raw[start+1 : end]) {
			return true
		}
	}
	return false
}

func providerIntentSecretKeyBytes(key []byte) bool {
	// Preserve strings.ToLower's Unicode semantics for non-ASCII keys, including
	// the Kelvin sign's lowercase k; common ASCII schema keys need no allocation.
	for _, c := range key {
		if c >= 0x80 {
			return providerIntentSecretKey(string(key))
		}
	}
	switch len(key) {
	case 6:
		return bytes.EqualFold(key, []byte("cookie")) || bytes.EqualFold(key, []byte("apikey"))
	case 7:
		return bytes.EqualFold(key, []byte("api_key")) || bytes.EqualFold(key, []byte("api-key"))
	case 9:
		return bytes.EqualFold(key, []byte("x-api-key"))
	case 10:
		return bytes.EqualFold(key, []byte("set-cookie"))
	case 12:
		return bytes.EqualFold(key, []byte("access_token"))
	case 13:
		return bytes.EqualFold(key, []byte("authorization")) || bytes.EqualFold(key, []byte("refresh_token"))
	case 19:
		return bytes.EqualFold(key, []byte("proxy-authorization"))
	}
	return false
}
