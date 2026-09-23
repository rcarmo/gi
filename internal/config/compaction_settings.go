package config

import (
	"encoding/json"
	"errors"
	"fmt"
)

var ErrCompactionPolicyInvalid = errors.New("invalid compaction policy")

const MaxCompactionTokens = 16_777_216

type CompactionPolicyEdit struct {
	Enabled          bool `json:"enabled"`
	ContextWindow    int  `json:"context_window"`
	ReserveTokens    int  `json:"reserve_tokens"`
	KeepRecentTokens int  `json:"keep_recent_tokens"`
	ThresholdTokens  int  `json:"threshold_tokens"`
}

type CompactionPolicySnapshot struct {
	Policy   CompactionSettings `json:"policy"`
	Revision string             `json:"revision"`
}

func validateCompactionPolicy(p CompactionPolicyEdit) error {
	if p.ContextWindow < 1 || p.ContextWindow > MaxCompactionTokens || p.ReserveTokens < 1 || p.ReserveTokens >= p.ContextWindow || p.KeepRecentTokens < 1 || p.KeepRecentTokens > p.ThresholdTokens || p.ThresholdTokens < 1 || p.ThresholdTokens > p.ContextWindow-p.ReserveTokens {
		return fmt.Errorf("%w: positive whole-number budgets required; context at most %d, reserve below context, keep recent <= threshold <= context minus reserve", ErrCompactionPolicyInvalid, MaxCompactionTokens)
	}
	return nil
}

func compactionPolicySnapshot(d settingsDocument) (CompactionPolicySnapshot, error) {
	// Match Load: absent settings file enables compaction, whereas an existing
	// settings document with no compaction section has zero-valued Enabled=false.
	p := CompactionSettings{Enabled: d.Revision == "missing"}
	if raw, ok := d.Values["compaction"]; ok {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw, &object); err != nil || object == nil {
			return CompactionPolicySnapshot{}, errors.New("compaction must be a JSON object")
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return CompactionPolicySnapshot{}, errors.New("compaction fields have invalid types")
		}
	}
	applyCompactionDefaults(&p)
	return CompactionPolicySnapshot{Policy: p, Revision: d.Revision}, nil
}

func ReadCompactionPolicy(workspace string) (CompactionPolicySnapshot, error) {
	d, err := readPiSettings(workspace)
	if err != nil {
		return CompactionPolicySnapshot{}, err
	}
	return compactionPolicySnapshot(d)
}

func SaveCompactionPolicy(workspace, revision string, p CompactionPolicyEdit) (CompactionPolicySnapshot, error) {
	if err := validateCompactionPolicy(p); err != nil {
		return CompactionPolicySnapshot{}, err
	}
	if revision == "" {
		return CompactionPolicySnapshot{}, ErrSettingsConflict
	}
	d, err := updatePiSettings(workspace, revision, func(d *settingsDocument) error {
		// Validate the stored section before editing, never replace malformed config.
		if _, err := compactionPolicySnapshot(*d); err != nil {
			return err
		}
		section := map[string]json.RawMessage{}
		if raw, ok := d.Values["compaction"]; ok {
			if err := json.Unmarshal(raw, &section); err != nil {
				return err
			}
		}
		raw, _ := json.Marshal(p)
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(raw, &fields)
		for key, value := range fields {
			section[key] = value
		}
		d.Values["compaction"], _ = json.Marshal(section)
		return nil
	})
	if err != nil {
		return CompactionPolicySnapshot{}, err
	}
	return compactionPolicySnapshot(d)
}
