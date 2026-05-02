package session

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const sessionKeyV1Prefix = "sk_v1_"

func BuildOpaqueSessionKey(alias string) string {
	normalized := strings.TrimSpace(strings.ToLower(alias))
	if normalized == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(normalized))
	return sessionKeyV1Prefix + hex.EncodeToString(sum[:])
}

func IsOpaqueSessionKey(key string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(key)), sessionKeyV1Prefix)
}

func CanonicalScopeSignature(scope SessionScope) string {
	parts := []string{
		fmt.Sprintf("v=%d", scope.Version),
		fmt.Sprintf("agent=%s", strings.TrimSpace(strings.ToLower(scope.AgentID))),
		fmt.Sprintf("channel=%s", strings.TrimSpace(strings.ToLower(scope.Channel))),
		fmt.Sprintf("account=%s", strings.TrimSpace(strings.ToLower(scope.Account))),
	}
	for _, dimension := range scope.Dimensions {
		dimension = strings.TrimSpace(strings.ToLower(dimension))
		if dimension == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", dimension, strings.TrimSpace(strings.ToLower(scope.Values[dimension]))))
	}
	return strings.Join(parts, "|")
}

func BuildSessionKey(scope SessionScope) string {
	return BuildOpaqueSessionKey(CanonicalScopeSignature(scope))
}
