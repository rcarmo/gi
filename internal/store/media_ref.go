package store

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type MediaRef struct {
	ID          string `json:"id"`
	MediaID     int64  `json:"media_id"`
	SessionID   string `json:"session_id,omitempty"`
	Filename    string `json:"filename,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	Size        int    `json:"size,omitempty"`
	SHA256      string `json:"sha256,omitempty"`
	Source      string `json:"source,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

func MediaRefID(mediaID int64) string {
	if mediaID <= 0 {
		return ""
	}
	return "media:" + strconv.FormatInt(mediaID, 10)
}

func ParseMediaRefID(value string) (int64, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if strings.HasPrefix(strings.ToLower(value), "media:") {
		value = strings.TrimSpace(value[len("media:"):])
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func NormalizeMediaReferences(raw any) ([]MediaRef, error) {
	if raw == nil {
		return nil, nil
	}
	var refs []MediaRef
	add := func(ref MediaRef) error {
		if ref.MediaID <= 0 {
			parsed, ok := ParseMediaRefID(ref.ID)
			if !ok {
				return fmt.Errorf("invalid media reference: %q", ref.ID)
			}
			ref.MediaID = parsed
		}
		if ref.ID == "" {
			ref.ID = MediaRefID(ref.MediaID)
		}
		refs = append(refs, ref)
		return nil
	}
	switch v := raw.(type) {
	case []MediaRef:
		for _, ref := range v {
			if err := add(ref); err != nil {
				return nil, err
			}
		}
	case []string:
		for _, value := range v {
			id, ok := ParseMediaRefID(value)
			if !ok {
				return nil, fmt.Errorf("invalid media reference: %q", value)
			}
			refs = append(refs, MediaRef{ID: MediaRefID(id), MediaID: id})
		}
	case []any:
		for _, item := range v {
			switch ref := item.(type) {
			case string:
				id, ok := ParseMediaRefID(ref)
				if !ok {
					return nil, fmt.Errorf("invalid media reference: %q", ref)
				}
				refs = append(refs, MediaRef{ID: MediaRefID(id), MediaID: id})
			case map[string]any:
				mr := MediaRef{}
				if value, _ := ref["id"].(string); value != "" {
					mr.ID = value
				}
				mr.MediaID = int64FromAny(ref["media_id"])
				mr.SessionID, _ = ref["session_id"].(string)
				mr.Filename, _ = ref["filename"].(string)
				mr.ContentType, _ = ref["content_type"].(string)
				mr.Size = int(int64FromAny(ref["size"]))
				mr.SHA256, _ = ref["sha256"].(string)
				mr.Source, _ = ref["source"].(string)
				mr.CreatedAt, _ = ref["created_at"].(string)
				if err := add(mr); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("invalid media reference type %T", item)
			}
		}
	default:
		return nil, fmt.Errorf("invalid media reference list type %T", raw)
	}
	return refs, nil
}

func int64FromAny(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case int32:
		return int64(n)
	case float64:
		return int64(n)
	case float32:
		return int64(n)
	case json.Number:
		id, _ := strconv.ParseInt(string(n), 10, 64)
		return id
	case string:
		id, _ := strconv.ParseInt(strings.TrimSpace(n), 10, 64)
		return id
	default:
		return 0
	}
}
