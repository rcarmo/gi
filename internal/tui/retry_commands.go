package tui

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/rcarmo/gi/internal/store"
)

const retryUsage = "retry: /retry [page] | check <id> | run <id> | release <id> <token>"
const retryPageSize = 6

func retryState(f *store.TurnFailure) string {
	if f.ResolutionState == "retry_pending" {
		if f.RetryAdmissionVersion == 1 {
			return "pending (check/release)"
		}
		return "pending (legacy; blocked)"
	}
	if f.ResolutionState != "" {
		return f.ResolutionState
	}
	if f.HoldState != "none" {
		return "held"
	}
	return "not held"
}
func retryPreview(value string) string {
	return selectorText(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value), 40)
}

func (c *chatTUI) retryCommand(fields []string) []string {
	if c.store == nil || c.engine == nil || c.sessionID == "" {
		return []string{"retry: active session required"}
	}
	ctx := context.Background()
	sessionID := c.sessionID
	page := 1
	if len(fields) > 1 {
		switch fields[1] {
		case "check", "run", "release":
			action := fields[1]
			want := 3
			if action == "release" {
				want = 4
			}
			if len(fields) != want {
				return []string{retryUsage}
			}
			id := fields[2]
			failure, err := c.engine.HeldRetryStatus(ctx, sessionID, id)
			if err != nil {
				return []string{"retry: check failed: " + err.Error()}
			}
			switch action {
			case "check":
				lines := []string{"retry: " + id + " · " + retryState(failure)}
				if failure.ResolvedTurnID != "" {
					lines = append(lines, "  admitted: "+failure.ResolvedTurnID)
				}
				if failure.ResolutionState == "retry_pending" {
					if failure.RetryAdmissionVersion == 1 {
						lines = append(lines, "  token: "+failure.RetryAdmissionToken, "  release only cancels unadmitted reservation; no submit")
					} else {
						lines = append(lines, "  legacy admission unknown; release refused, no resend")
					}
				}
				return lines
			case "release":
				if err = c.engine.ReleaseHeldRetryInSession(ctx, sessionID, id, fields[3]); err != nil {
					return []string{"retry: release failed; check ID/token", "  " + err.Error()}
				}
				return []string{"retry: released " + id, "  no work submitted; held for explicit action"}
			case "run":
				result, err := c.engine.RetryHeldTurnInSession(ctx, sessionID, id, "Explicit terminal retry")
				if err != nil {
					if errors.Is(err, store.ErrRetryPending) {
						return []string{"retry: admission pending; no resend", "  /retry check " + id}
					}
					return []string{"retry: failed; check native status before retrying: " + err.Error()}
				}
				c.queueSnapshot = nil
				return []string{"retry: admitted " + result.TurnID, "  " + result.Status + "; check/repeat returns stored ID"}
			}
		default:
			if len(fields) != 2 {
				return []string{retryUsage}
			}
			value, err := strconv.Atoi(fields[1])
			if err != nil || value < 1 || value > (int(^uint(0)>>1)-retryPageSize)/retryPageSize {
				return []string{retryUsage}
			}
			page = value
		}
	}
	items, err := c.store.ListHeldTurnFailures(ctx, sessionID, (page-1)*retryPageSize, retryPageSize+1)
	if err != nil {
		return []string{"retry: read failed: " + err.Error()}
	}
	if len(items) == 0 {
		if page > 1 {
			return []string{"retry: page out of range; /retry to refresh"}
		}
		return []string{"retry: no held failures", retryUsage}
	}
	lines := []string{fmt.Sprintf("retry: held failures · page %d", page)}
	for _, item := range items[:min(len(items), retryPageSize)] {
		lines = append(lines, "  "+item.TurnID+"  "+retryState(&item)+"  "+retryPreview(item.Summary))
	}
	if len(items) > retryPageSize {
		lines = append(lines, fmt.Sprintf("  more: /retry %d", page+1))
	}
	return append(lines, retryUsage)
}
