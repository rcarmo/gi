package config

import "strings"

const (
	defaultHookTimeoutMS = 1500
	hookPolicyError      = "error"
	hookPolicyContinue   = "continue"
)

func applyHookDefaults(cfg *HookSettings) {
	if cfg == nil {
		return
	}
	if cfg.TimeoutMS <= 0 {
		cfg.TimeoutMS = defaultHookTimeoutMS
	}
	cfg.OnError = NormalizeHookPolicy(cfg.OnError, hookPolicyError)
	cfg.OnTimeout = NormalizeHookPolicy(cfg.OnTimeout, hookPolicyContinue)
}

func NormalizeHookPolicy(value, fallback string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case hookPolicyError:
		return hookPolicyError
	case hookPolicyContinue:
		return hookPolicyContinue
	default:
		return fallback
	}
}
