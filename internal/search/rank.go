package search

import "sort"

// MergeAndRank combines FTS and vector hits by chunk id and computes a simple
// weighted hybrid score suitable for a first implementation.
func MergeAndRank(ftsHits, vecHits []SearchHit, limit int) []SearchHit {
	merged := map[int64]SearchHit{}
	for _, hit := range ftsHits {
		current := merged[hit.ChunkID]
		if current.ChunkID == 0 {
			current = hit
		} else {
			current = mergeHitMetadata(current, hit)
		}
		current.FTSScore = hit.FTSScore
		merged[hit.ChunkID] = current
	}
	for _, hit := range vecHits {
		current := merged[hit.ChunkID]
		if current.ChunkID == 0 {
			current = hit
		} else {
			current = mergeHitMetadata(current, hit)
		}
		current.VecScore = hit.VecScore
		merged[hit.ChunkID] = current
	}
	result := make([]SearchHit, 0, len(merged))
	for _, hit := range merged {
		hit.FinalScore = 0.55*hit.VecScore + 0.35*hit.FTSScore + headingBoost(hit) + pathBoost(hit)
		result = append(result, hit)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].FinalScore == result[j].FinalScore {
			if result[i].Path == result[j].Path {
				return result[i].ChunkID < result[j].ChunkID
			}
			return result[i].Path < result[j].Path
		}
		return result[i].FinalScore > result[j].FinalScore
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

func mergeHitMetadata(a, b SearchHit) SearchHit {
	if a.DocumentID == 0 {
		a.DocumentID = b.DocumentID
	}
	if a.Path == "" {
		a.Path = b.Path
	}
	if a.Language == "" {
		a.Language = b.Language
	}
	if a.Heading == "" {
		a.Heading = b.Heading
	}
	if a.Content == "" {
		a.Content = b.Content
	}
	if a.StartLine == 0 {
		a.StartLine = b.StartLine
	}
	if a.EndLine == 0 {
		a.EndLine = b.EndLine
	}
	return a
}

func headingBoost(hit SearchHit) float64 {
	if hit.Heading == "" {
		return 0
	}
	return 0.05
}

func pathBoost(hit SearchHit) float64 {
	if hit.Path == "" {
		return 0
	}
	return 0.05
}
