package routing

type Classifier interface{ Score(f Features) float64 }

type RuleClassifier struct{}

func (c *RuleClassifier) Score(f Features) float64 {
	if f.HasAttachments {
		return 1.0
	}
	var score float64
	switch {
	case f.TokenEstimate > 200:
		score += 0.35
	case f.TokenEstimate > 50:
		score += 0.15
	}
	if f.CodeBlockCount > 0 {
		score += 0.40
	}
	switch {
	case f.RecentToolCalls > 3:
		score += 0.25
	case f.RecentToolCalls > 0:
		score += 0.10
	}
	if f.ConversationDepth > 10 {
		score += 0.10
	}
	if score > 1.0 {
		score = 1.0
	}
	return score
}
