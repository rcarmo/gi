package config

// AgentConfig describes one routable agent identity.
type AgentConfig struct {
	ID           string   `json:"id"`
	Name         string   `json:"name,omitempty"`
	Default      bool     `json:"default,omitempty"`
	Model        string   `json:"model,omitempty"`
	LightModel   string   `json:"light_model,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// DispatchSelector matches normalized inbound context fields.
type DispatchSelector struct {
	Channel   string `json:"channel,omitempty"`
	Account   string `json:"account,omitempty"`
	Space     string `json:"space,omitempty"`
	Chat      string `json:"chat,omitempty"`
	Topic     string `json:"topic,omitempty"`
	Sender    string `json:"sender,omitempty"`
	Mentioned *bool  `json:"mentioned,omitempty"`
}

// DispatchRule routes inbound traffic to an agent and optional session dimensions.
type DispatchRule struct {
	Name              string           `json:"name,omitempty"`
	Agent             string           `json:"agent,omitempty"`
	When              DispatchSelector `json:"when,omitempty"`
	SessionDimensions []string         `json:"session_dimensions,omitempty"`
}

// DispatchConfig holds ordered routing rules.
type DispatchConfig struct {
	Rules []DispatchRule `json:"rules,omitempty"`
}

// AgentsConfig holds known agents and dispatch rules.
type AgentsConfig struct {
	List     []AgentConfig   `json:"list,omitempty"`
	Dispatch *DispatchConfig `json:"dispatch,omitempty"`
}

// SessionConfig controls route-to-session isolation policy.
type SessionConfig struct {
	Dimensions    []string            `json:"dimensions,omitempty"`
	IdentityLinks map[string][]string `json:"identity_links,omitempty"`
}

// ModelRoutingConfig controls light-vs-primary model selection.
type ModelRoutingConfig struct {
	Enabled    bool    `json:"enabled,omitempty"`
	LightModel string  `json:"light_model,omitempty"`
	Threshold  float64 `json:"threshold,omitempty"`
}
