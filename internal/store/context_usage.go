package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

// ContextMeasurement is the latest observed provider request, not a sum of
// iterations or a current tokenizer estimate. Output is excluded from Tokens.
type ContextMeasurement struct {
	Tokens     int    `json:"tokens"`
	Input      int    `json:"input"`
	CacheRead  int    `json:"cache_read"`
	CacheWrite int    `json:"cache_write"`
	Model      string `json:"model"`
	TurnID     string `json:"turn_id"`
	Iteration  int    `json:"iteration"`
	ObservedAt string `json:"observed_at"`
}

// Only explicit context.measured events qualify. Historical inference.finished
// totals are cumulative, so treating them as occupancy would inflate the meter.
func (s *Store) LatestContextMeasurement(ctx context.Context, sessionID string) (*ContextMeasurement, error) {
	var raw, turnID, at string
	err := s.db.QueryRowContext(ctx, `select payload_json, turn_id, created_at from turn_events where session_id = ? and event_type = 'context.measured' order by rowid desc limit 1`, sessionID).Scan(&raw, &turnID, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var measurement ContextMeasurement
	if err := json.Unmarshal([]byte(raw), &measurement); err != nil {
		return nil, err
	}
	if measurement.Tokens < 0 || measurement.Input < 0 || measurement.CacheRead < 0 || measurement.CacheWrite < 0 {
		return nil, nil
	}
	measurement.TurnID, measurement.ObservedAt = turnID, at
	return &measurement, nil
}
