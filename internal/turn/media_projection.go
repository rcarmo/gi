package turn

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

var providerSafeImageTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
	"image/gif":  true,
}

func (e *Engine) resolveMediaReferences(ctx context.Context, sessionID string, raw any) ([]store.MediaRef, error) {
	refs, err := store.NormalizeMediaReferences(raw)
	if err != nil {
		return nil, err
	}
	for i := range refs {
		media, err := e.store.GetMedia(ctx, refs[i].MediaID)
		if err != nil {
			return nil, fmt.Errorf("resolve media %s: %w", refs[i].ID, err)
		}
		if media.SessionID != sessionID {
			return nil, fmt.Errorf("media %s does not belong to session", refs[i].ID)
		}
		refs[i].ID = store.MediaRefID(media.ID)
		refs[i].MediaID = media.ID
		refs[i].SessionID = media.SessionID
		refs[i].Filename = media.Filename
		refs[i].ContentType = media.ContentType
		refs[i].Size = media.OriginalSize
		refs[i].SHA256, _ = media.Metadata["sha256"].(string)
		refs[i].Source, _ = media.Metadata["source"].(string)
		refs[i].CreatedAt = media.CreatedAt
	}
	return refs, nil
}

func (r *sessionRunner) userMessageWithProviderSafeMedia(ctx context.Context, sessionID, content string, payload map[string]any) goai.Message {
	blocks := []goai.ContentBlock{{Type: "text", Text: content}}
	refs, err := store.NormalizeMediaReferences(payload["media"])
	if err != nil || len(refs) == 0 {
		return goai.Message{Role: goai.RoleUser, Content: blocks}
	}
	projected := 0
	for _, ref := range refs {
		media, raw, err := r.store.GetMediaContent(ctx, ref.MediaID)
		if err != nil || media.SessionID != sessionID {
			continue
		}
		mimeType := strings.ToLower(strings.TrimSpace(media.ContentType))
		if !providerSafeImageTypes[mimeType] || media.OriginalSize > 10<<20 {
			continue
		}
		blocks = append(blocks, goai.ContentBlock{Type: "image", MimeType: mimeType, Data: base64.StdEncoding.EncodeToString(raw)})
		projected++
	}
	if projected == 0 && len(refs) > 0 {
		blocks[0].Text = appendMediaPlaceholder(content)
	}
	return goai.Message{Role: goai.RoleUser, Content: blocks}
}

func appendMediaPlaceholder(content string) string {
	if strings.TrimSpace(content) == "" {
		return "[user provided media attachments]"
	}
	return content + "\n\n[media attachments included]"
}
