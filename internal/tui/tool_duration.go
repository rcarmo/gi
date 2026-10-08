package tui

import "math"

// recordedToolDuration accepts only a non-negative integral measurement that
// fits time.Duration in milliseconds. Missing/invalid legacy data stays unknown.
func recordedToolDuration(value any) *int64 {
	const maxMS = int64(math.MaxInt64 / 1000000)
	var ms int64
	switch v := value.(type) {
	case int:
		ms = int64(v)
	case int64:
		ms = v
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > float64(maxMS) || math.Trunc(v) != v {
			return nil
		}
		ms = int64(v)
	default:
		return nil
	}
	if ms < 0 || ms > maxMS {
		return nil
	}
	return &ms
}
