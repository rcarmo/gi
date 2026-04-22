package web

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type metricsCollector struct {
	mu             sync.Mutex
	prevIdleTime   uint64
	prevTotalTime  uint64
	cpuPercent     float64
	cpuSeries      []float64
	ramSeries      []float64
	swapSeries     []float64
	rssSeries      []float64
	maxSamples     int
	sampleInterval time.Duration
	lastSample     time.Time
}

var collector = &metricsCollector{
	maxSamples:     30,
	sampleInterval: 2 * time.Second,
}

func (s *Server) handleSystemMetrics(w http.ResponseWriter, r *http.Request) {
	snapshot := collector.sample()
	writeJSON(w, http.StatusOK, snapshot)
}

func (c *metricsCollector) sample() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()

	cpu := c.readCPU()
	ram := readRAM()
	swap := readSwap()
	rss := readRSS()

	c.cpuSeries = pushSample(c.cpuSeries, cpu, c.maxSamples)
	c.ramSeries = pushSample(c.ramSeries, ram.percent, c.maxSamples)
	if swap.percent >= 0 {
		c.swapSeries = pushSample(c.swapSeries, swap.percent, c.maxSamples)
	}
	c.rssSeries = pushSample(c.rssSeries, float64(rss), c.maxSamples)
	c.lastSample = time.Now()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]any{
		"cpu_percent":                    cpu,
		"ram_percent":                    ram.percent,
		"swap_percent":                   nilIfNeg(swap.percent),
		"cpu_series":                     c.cpuSeries,
		"ram_series":                     c.ramSeries,
		"swap_series":                    c.swapSeries,
		"process_rss_series_bytes":       c.rssSeries,
		"process_heap_used_series_bytes": []float64{float64(memStats.HeapAlloc)},
		"swap_total_bytes":               swap.totalBytes,
		"swap_used_bytes":                swap.usedBytes,
		"sample_interval_ms":             int(c.sampleInterval.Milliseconds()),
		"platform":                       runtime.GOOS,
		"process_memory": map[string]any{
			"rss_bytes":        rss,
			"heap_total_bytes": memStats.HeapSys,
			"heap_used_bytes":  memStats.HeapAlloc,
		},
	}
}

func (c *metricsCollector) readCPU() float64 {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return c.cpuPercent
	}
	line := strings.SplitN(string(data), "\n", 2)[0]
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return c.cpuPercent
	}
	var idle, total uint64
	for i, f := range fields[1:] {
		v, _ := strconv.ParseUint(f, 10, 64)
		total += v
		if i == 3 {
			idle = v
		}
	}
	if c.prevTotalTime > 0 {
		deltaTotal := total - c.prevTotalTime
		deltaIdle := idle - c.prevIdleTime
		if deltaTotal > 0 {
			c.cpuPercent = round2(100.0 * float64(deltaTotal-deltaIdle) / float64(deltaTotal))
		}
	}
	c.prevTotalTime = total
	c.prevIdleTime = idle
	return c.cpuPercent
}

type memInfo struct {
	percent    float64
	totalBytes uint64
	usedBytes  uint64
}

func readRAM() memInfo {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return memInfo{}
	}
	m := parseMemInfo(string(data))
	total := m["MemTotal"]
	available := m["MemAvailable"]
	if total == 0 {
		return memInfo{totalBytes: total * 1024}
	}
	used := total - available
	return memInfo{
		percent:    round2(100.0 * float64(used) / float64(total)),
		totalBytes: total * 1024,
		usedBytes:  used * 1024,
	}
}

func readSwap() memInfo {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return memInfo{percent: -1}
	}
	m := parseMemInfo(string(data))
	total := m["SwapTotal"]
	free := m["SwapFree"]
	if total == 0 {
		return memInfo{percent: -1}
	}
	used := total - free
	return memInfo{
		percent:    round2(100.0 * float64(used) / float64(total)),
		totalBytes: total * 1024,
		usedBytes:  used * 1024,
	}
}

func readRSS() uint64 {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/statm", os.Getpid()))
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return 0
	}
	pages, _ := strconv.ParseUint(fields[1], 10, 64)
	return pages * 4096
}

func parseMemInfo(data string) map[string]uint64 {
	m := make(map[string]uint64)
	for _, line := range strings.Split(data, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.TrimSuffix(val, " kB")
		v, _ := strconv.ParseUint(strings.TrimSpace(val), 10, 64)
		m[key] = v
	}
	return m
}

func pushSample(series []float64, value float64, max int) []float64 {
	series = append(series, value)
	if len(series) > max {
		series = series[len(series)-max:]
	}
	return series
}

func round2(v float64) float64 {
	return float64(int(v*100)) / 100
}

func nilIfNeg(v float64) any {
	if v < 0 {
		return nil
	}
	return v
}
