package turn

import "github.com/rcarmo/gi/internal/connectivity"

// Connectivity returns the engine-wide connectivity registry. Routes are
// transport-neutral; web/socket adapters dispatch through this registry.
func (e *Engine) Connectivity() *connectivity.Registry { return e.connectivity }
