package web

import "testing"

func TestMetricsCollectorSampleReturnsDetachedSeriesCopies(t *testing.T) {
	c := &metricsCollector{
		cpuPercent:     12.5,
		cpuSeries:      []float64{1, 2},
		ramSeries:      []float64{3, 4},
		swapSeries:     []float64{5, 6},
		rssSeries:      []float64{7, 8},
		maxSamples:     8,
		sampleInterval: 2,
	}

	snapshot := c.sample()
	cpuSeries, ok := snapshot["cpu_series"].([]float64)
	if !ok {
		t.Fatalf("expected cpu_series slice, got %#v", snapshot["cpu_series"])
	}
	ramSeries, ok := snapshot["ram_series"].([]float64)
	if !ok {
		t.Fatalf("expected ram_series slice, got %#v", snapshot["ram_series"])
	}
	swapSeries, ok := snapshot["swap_series"].([]float64)
	if !ok {
		t.Fatalf("expected swap_series slice, got %#v", snapshot["swap_series"])
	}
	rssSeries, ok := snapshot["process_rss_series_bytes"].([]float64)
	if !ok {
		t.Fatalf("expected rss series slice, got %#v", snapshot["process_rss_series_bytes"])
	}

	cpuSeries[0] = 999
	ramSeries[0] = 999
	swapSeries[0] = 999
	rssSeries[0] = 999

	if c.cpuSeries[0] == 999 || c.ramSeries[0] == 999 || c.swapSeries[0] == 999 || c.rssSeries[0] == 999 {
		t.Fatalf("expected snapshot mutation not to affect collector state: %#v %#v %#v %#v", c.cpuSeries, c.ramSeries, c.swapSeries, c.rssSeries)
	}
}
