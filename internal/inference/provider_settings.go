package inference

import (
	"encoding/json"
	"fmt"
	"sort"
	"unicode"
	"unicode/utf8"
)

type ProviderCredentialInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Stored      bool   `json:"stored"`
	HasMaterial bool   `json:"has_material"`
	Kind        string `json:"kind"`
	Editable    bool   `json:"editable"`
}
type ProviderSettingsSnapshot struct {
	Providers []ProviderCredentialInfo `json:"providers"`
	Revision  string                   `json:"revision"`
	Scope     string                   `json:"scope"`
}

// Metadata never authenticates with a remote provider or returns credential values.
func ReadProviderSettings() (ProviderSettingsSnapshot, error) {
	d, err := readCredentials()
	if err != nil {
		return ProviderSettingsSnapshot{}, err
	}
	ids := map[string]bool{"openai": true, "anthropic": true}
	for id := range d.entries {
		ids[id] = true
	}
	snapshot := ProviderSettingsSnapshot{Revision: d.revision, Scope: "service-user"}
	for id := range ids {
		if !utf8.ValidString(id) || len(id) > 128 {
			return snapshot, fmt.Errorf("invalid provider identifier in credential store")
		}
		for _, r := range id {
			if unicode.IsControl(r) {
				return snapshot, fmt.Errorf("invalid provider identifier in credential store")
			}
		}
		raw, stored := d.entries[id]
		var entry authEntry
		if stored && json.Unmarshal(raw, &entry) != nil {
			return snapshot, fmt.Errorf("invalid credential metadata")
		}
		kind := "none"
		if stored {
			kind = "credential"
			if entry.Type == "oauth" {
				kind = "oauth"
			} else if editableKeyEntry(raw) {
				kind = "api_key"
			}
		}
		snapshot.Providers = append(snapshot.Providers, ProviderCredentialInfo{ID: id, Name: providerName(id), Stored: stored, HasMaterial: entry.APIKey != "" || entry.Access != "" || entry.Refresh != "" || entry.Token != "", Kind: kind, Editable: APIKeyProvider(id) && editableKeyEntry(raw)})
	}
	sort.Slice(snapshot.Providers, func(i, j int) bool { return snapshot.Providers[i].ID < snapshot.Providers[j].ID })
	return snapshot, nil
}
