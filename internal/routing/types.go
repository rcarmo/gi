package routing

// SessionPolicy is the routing-selected session isolation policy.
type SessionPolicy struct {
	Dimensions    []string
	IdentityLinks map[string][]string
}

// ResolvedRoute is the route decision for an inbound message.
type ResolvedRoute struct {
	AgentID       string
	Channel       string
	AccountID     string
	SessionPolicy SessionPolicy
	MatchedBy     string
}

// InboundContext is gi's normalized inbound message view for routing.
type InboundContext struct {
	Channel   string
	Account   string
	SpaceType string
	SpaceID   string
	ChatType  string
	ChatID    string
	TopicID   string
	SenderID  string
	Mentioned bool
	Prompt    string
}
