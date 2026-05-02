package routing

const defaultThreshold = 0.35

type Router struct {
	cfg        ModelRoutingConfig
	classifier Classifier
}

func NewRouter(cfg ModelRoutingConfig) *Router {
	if cfg.Threshold <= 0 {
		cfg.Threshold = defaultThreshold
	}
	return &Router{cfg: cfg, classifier: &RuleClassifier{}}
}

func (r *Router) SelectModel(msg string, history []HistoryMessage, primaryModel string) (model string, usedLight bool, score float64) {
	features := ExtractFeatures(msg, history)
	score = r.classifier.Score(features)
	if r.cfg.Enabled && r.cfg.LightModel != "" && score < r.cfg.Threshold {
		return r.cfg.LightModel, true, score
	}
	return primaryModel, false, score
}
