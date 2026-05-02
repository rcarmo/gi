package turn

// executeToolsTool preserves the historical package-level helper used by tests.
func executeToolsTool(args map[string]any) (string, error) {
	return New(nil).executeToolsTool(args)
}
