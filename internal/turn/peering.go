package turn

import "github.com/rcarmo/gi/internal/peering"

func (e *Engine) PeeringStatus() peering.Status {
	if e.peering == nil {
		return peering.Status{Backend: "tsnet", State: "unavailable"}
	}
	return e.peering.Status()
}
