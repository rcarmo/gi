package config

func applyCompactionDefaults(c *CompactionSettings) {
	if c.ContextWindow <= 0 {
		c.ContextWindow = 128000
	}
	if c.ReserveTokens <= 0 {
		c.ReserveTokens = 20000
	}
	if c.KeepRecentTokens <= 0 {
		c.KeepRecentTokens = 20000
	}
	if c.ThresholdTokens <= 0 {
		threshold := c.ContextWindow - c.ReserveTokens
		if threshold <= 0 {
			threshold = 100000
		}
		c.ThresholdTokens = threshold
	}
	if c.Strategy == "" {
		c.Strategy = "default"
	}
}
