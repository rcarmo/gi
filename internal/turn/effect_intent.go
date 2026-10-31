package turn

import (
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
	var owned any
	if err = json.Unmarshal(raw, &owned); err != nil {
		return nil, err
	}
	var sanitize func(any)
	sanitize = func(value any) {
		switch object := value.(type) {
		case map[string]any:
			for key, child := range object {
				switch strings.ToLower(key) {
				case "authorization", "proxy-authorization", "api_key", "apikey", "api-key", "x-api-key", "access_token", "refresh_token", "cookie", "set-cookie":
					delete(object, key)
				default:
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
