package queue

const (
	inboundDispatcherLeaseNamespace = "runtime_leases"
	inboundDispatcherLeaseKey       = "inbound_dispatcher"
	defaultClaimedByWorker          = "worker"

	statusQueued    = "queued"
	statusRetry     = "retry"
	statusClaimed   = "claimed"
	statusCompleted = "completed"
	statusFailed    = "failed"
	statusDiscarded = "discarded"
)
