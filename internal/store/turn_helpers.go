package store

import (
	"fmt"
	"strings"
)

func RuntimeTurnPhaseForStatus(status string) string {
	switch status {
	case "queued":
		return "queued"
	case "running":
		return "setup"
	case "cancelling":
		return "cancelling"
	case "completed":
		return "completed"
	case "failed":
		return "failed"
	case "aborted", "cancelled":
		return "aborted"
	default:
		return status
	}
}

func NormalizeSubTurnDeliveryMode(mode string) (string, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		return "sync", nil
	}
	switch mode {
	case "sync", "async":
		return mode, nil
	default:
		return "", fmt.Errorf("invalid subturn delivery mode: %s", mode)
	}
}
