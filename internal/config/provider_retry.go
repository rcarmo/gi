package config

// ProviderRetrySettings follows Pi's retry settings. Pointer fields distinguish
// explicit false/zero from missing values in programmatically built configs.
type ProviderRetrySettings struct {
	Enabled         *bool `json:"enabled,omitempty"`
	MaxRetries      *int  `json:"maxRetries,omitempty"`
	BaseDelayMS     *int  `json:"baseDelayMs,omitempty"`
	MaxAgentDelayMS *int  `json:"maxAgentDelayMs,omitempty"`
}

type ProviderRetryPolicy struct {
	Enabled                             bool
	MaxRetries, BaseDelayMS, MaxDelayMS int
}

func (s ProviderRetrySettings) Policy() ProviderRetryPolicy {
	p := ProviderRetryPolicy{true, 3, 2000, 60000}
	if s.Enabled != nil {
		p.Enabled = *s.Enabled
	}
	if s.MaxRetries != nil && *s.MaxRetries >= 0 {
		p.MaxRetries = *s.MaxRetries
	}
	if s.BaseDelayMS != nil && *s.BaseDelayMS >= 0 {
		p.BaseDelayMS = *s.BaseDelayMS
	}
	if s.MaxAgentDelayMS != nil && *s.MaxAgentDelayMS >= 0 {
		p.MaxDelayMS = *s.MaxAgentDelayMS
	}
	// Keep malformed or hostile settings bounded.
	if p.MaxRetries > 100 {
		p.MaxRetries = 100
	}
	if p.BaseDelayMS > 60000 {
		p.BaseDelayMS = 60000
	}
	if p.MaxDelayMS > 60000 {
		p.MaxDelayMS = 60000
	}
	return p
}
