// Package secrets defines the host-independent named-secret resolution contract.
package secrets

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unicode"
)

// Resolver resolves a named secret without exposing backend details to callers.
type Resolver interface {
	Resolve(context.Context, string) (string, error)
}

// ResolverFunc adapts a function to Resolver.
type ResolverFunc func(context.Context, string) (string, error)

func (f ResolverFunc) Resolve(ctx context.Context, name string) (string, error) {
	return f(ctx, name)
}

// EnvResolver resolves Piclaw-style keychain names from their injected
// environment variables. '/', '-', and '.' (and any other non-alphanumeric
// rune) become underscores and the name is uppercased.
type EnvResolver struct{}

func (EnvResolver) Resolve(_ context.Context, name string) (string, error) {
	env := EnvName(name)
	if env == "" {
		return "", fmt.Errorf("secret name is required")
	}
	value := os.Getenv(env)
	if value == "" {
		return "", fmt.Errorf("secret %q is not available in %s", name, env)
	}
	return value, nil
}

// EnvName converts a keychain entry name to its injected environment name.
func EnvName(name string) string {
	name = strings.TrimSpace(name)
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(unicode.ToUpper(r))
		} else {
			b.WriteByte('_')
		}
	}
	return strings.Trim(b.String(), "_")
}
