package turn

func terminalPhaseForStatus(status string) string {
	switch status {
	case "completed":
		return "completed"
	case "failed":
		return "failed"
	case "aborted", "cancelled":
		return "aborted"
	case "cancelling":
		return "cancelling"
	case "queued":
		return "queued"
	default:
		return "running"
	}
}
