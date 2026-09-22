package inference

import (
	"context"
	"fmt"

	"github.com/rcarmo/gi/internal/store"
)

// SessionContextUsage exposes only measured values. Nil fields remain unknown.
// Capacity is catalogue metadata; it is not an invented token measurement.
func SessionContextUsage(ctx context.Context, s *store.Store, sessionID string, window int) (map[string]any, error) {
	measurement, err := s.LatestContextMeasurement(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	result := map[string]any{"tokens": nil, "contextWindow": nil, "percent": nil, "source": "unavailable", "measurement": nil}
	if window > 0 {
		result["contextWindow"] = window
	}
	if measurement != nil {
		result["tokens"] = measurement.Tokens
		result["source"] = "provider_request"
		result["measurement"] = measurement
		if window > 0 {
			result["percent"] = float64(measurement.Tokens) * 100 / float64(window)
		}
	}
	return result, nil
}

func CheckSessionModelContext(ctx context.Context, s *store.Store, sessionID string, option ModelOption) error {
	if option.ContextWindow <= 0 {
		return nil
	}
	measurement, err := s.LatestContextMeasurement(ctx, sessionID)
	if err != nil {
		return err
	}
	if measurement != nil && measurement.Tokens > option.ContextWindow {
		return fmt.Errorf("model %q context window %d is smaller than latest measured request (%d tokens)", option.Label, option.ContextWindow, measurement.Tokens)
	}
	return nil
}
