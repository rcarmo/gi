// Package scripting defines the bridge interface between gi's runtime and
// script engines (Joker/Clojure, QuickJS/JavaScript).
//
// The bridge exposes gi's internal state to scripts in a controlled way.
// Each engine implements the Runner interface; the bridge provides the
// host functions that scripts can call.
package scripting

import (
	"context"
	"fmt"
)

// Runner is the interface that script engines must implement.
type Runner interface {
	// Name returns the engine name (e.g. "joker", "quickjs").
	Name() string

	// Execute runs a script string and returns the result.
	Execute(ctx context.Context, script string, bridge *Bridge) (string, error)

	// ExecuteFile runs a script file and returns the result.
	ExecuteFile(ctx context.Context, path string, bridge *Bridge) (string, error)
}

// Bridge provides host functions that scripts can call to interact with
// gi's runtime state. It is the single point of contact between the
// script engine and the gi internals.
type Bridge struct {
	// SessionID is the active session.
	SessionID string

	// Funcs are the callable host functions exposed to scripts.
	Funcs BridgeFuncs
}

// BridgeFuncs defines the host functions available to scripts.
// Each function takes and returns generic types that the engine adapter
// marshals to/from the script language's native types.
type BridgeFuncs struct {
	// Session state
	GetSessionState func(ctx context.Context) (map[string]any, error)
	SetSessionState func(ctx context.Context, patch map[string]any) error

	// Messages
	ListMessages func(ctx context.Context, limit int) ([]map[string]any, error)
	AddMessage   func(ctx context.Context, role, content string) error

	// Turn events
	ListTurnEvents func(ctx context.Context, turnID string) ([]map[string]any, error)

	// Runtime config
	GetConfig func(ctx context.Context) (map[string]any, error)

	// Workspace files
	ReadFile  func(ctx context.Context, path string) (string, error)
	WriteFile func(ctx context.Context, path, content string) error
	ListDir   func(ctx context.Context, path string) ([]map[string]any, error)

	// Shell execution
	Exec func(ctx context.Context, command string) (string, error)

	// Logging
	Log func(ctx context.Context, level, message string)
}

// NewBridge creates a bridge with the given session ID and functions.
func NewBridge(sessionID string, funcs BridgeFuncs) *Bridge {
	return &Bridge{SessionID: sessionID, Funcs: funcs}
}

// Validate checks that all required bridge functions are set.
func (b *Bridge) Validate() error {
	if b.Funcs.GetSessionState == nil {
		return fmt.Errorf("bridge: GetSessionState not set")
	}
	if b.Funcs.ListMessages == nil {
		return fmt.Errorf("bridge: ListMessages not set")
	}
	if b.Funcs.GetConfig == nil {
		return fmt.Errorf("bridge: GetConfig not set")
	}
	if b.Funcs.ReadFile == nil {
		return fmt.Errorf("bridge: ReadFile not set")
	}
	if b.Funcs.Log == nil {
		return fmt.Errorf("bridge: Log not set")
	}
	return nil
}
