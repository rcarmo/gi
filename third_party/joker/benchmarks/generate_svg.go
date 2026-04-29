//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
)

type Benchmark struct {
	MSPerOp     float64 `json:"ms_per_op"`
	BytesPerOp  float64 `json:"bytes_per_op"`
	AllocsPerOp float64 `json:"allocs_per_op"`
}

type Series struct {
	ID         string               `json:"id"`
	Label      string               `json:"label"`
	Date       string               `json:"date"`
	Notes      []string             `json:"notes"`
	Benchmarks map[string]Benchmark `json:"benchmarks"`
}

type MapStoryCheckpoint struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	Date      string    `json:"date"`
	Notes     []string  `json:"notes"`
	Benchmark Benchmark `json:"benchmark"`
}

type MapStory struct {
	Label       string               `json:"label"`
	Description string               `json:"description"`
	Checkpoints []MapStoryCheckpoint `json:"checkpoints"`
}

type History struct {
	Metadata struct {
		Project string `json:"project"`
		Scope   string `json:"scope"`
		Host    string `json:"host"`
		Command string `json:"command"`
	} `json:"metadata"`
	Series         []Series `json:"series"`
	MapUpdateStory MapStory `json:"map_update_story"`
}

type OverviewRow struct {
	Label        string
	BaselineMS   float64
	LatestMS     float64
	DeltaPct     float64
	CurrentWidth int
}

type MapRow struct {
	Label    string
	MS       float64
	Width    int
	DeltaPct float64
	Kind     string
}

type SVGData struct {
	Title      string
	Subtitle   string
	Host       string
	Rows       []OverviewRow
	MapRows    []MapRow
	MapNote    string
	PanelWidth int
}

func main() {
	baseDir := "."
	if len(os.Args) > 1 {
		baseDir = os.Args[1]
	}
	jsonPath := filepath.Join(baseDir, "benchmark-history.json")
	svgPath := filepath.Join(baseDir, "benchmark-improvements.svg")

	data, err := os.ReadFile(jsonPath)
	must(err)

	var h History
	must(json.Unmarshal(data, &h))

	baseline := findSeries(h.Series, "baseline-stable")
	latest := findSeries(h.Series, "current")

	rows := []OverviewRow{
		makeOverviewRow("Arithmetic loop", baseline, latest, "arithmetic_loop"),
		makeOverviewRow("Recursive fib", baseline, latest, "recursive_fib"),
		makeOverviewRow("Word frequency", baseline, latest, "word_frequency"),
	}

	mapRows := makeMapRows(h.MapUpdateStory.Checkpoints)

	view := SVGData{
		Title:      "Joker benchmark improvements",
		Subtitle:   "Stable 5x benchmark checkpoints • lower is better • generated from benchmark-history.json",
		Host:       h.Metadata.Host,
		Rows:       rows,
		MapRows:    mapRows,
		MapNote:    "The map benchmark improved after `get`/`assoc` fast paths, but later structural experiments moved the latest value back up.",
		PanelWidth: 1104,
	}

	f, err := os.Create(svgPath)
	must(err)
	defer f.Close()

	tmpl := template.Must(template.New("svg").Funcs(template.FuncMap{
		"fmtMS":     fmtMS,
		"fmtPct":    fmtPct,
		"deltaCls":  deltaClass,
		"addY":      addY,
		"mapLabelY": mapLabelY,
		"mapTrackY": mapTrackY,
		"mapTextY":  mapTextY,
	}).Parse(svgTemplate))
	must(tmpl.Execute(f, view))
	fmt.Println("wrote", svgPath)
}

func findSeries(series []Series, id string) Series {
	for _, s := range series {
		if s.ID == id {
			return s
		}
	}
	panic("missing series: " + id)
}

func makeOverviewRow(label string, baseline, latest Series, key string) OverviewRow {
	b := baseline.Benchmarks[key]
	l := latest.Benchmarks[key]
	return OverviewRow{
		Label:        label,
		BaselineMS:   b.MSPerOp,
		LatestMS:     l.MSPerOp,
		DeltaPct:     pctChange(b.MSPerOp, l.MSPerOp),
		CurrentWidth: barWidth(l.MSPerOp, b.MSPerOp, 620),
	}
}

func makeMapRows(checkpoints []MapStoryCheckpoint) []MapRow {
	rows := make([]MapRow, 0, len(checkpoints))
	max := 0.0
	for _, cp := range checkpoints {
		if cp.Benchmark.MSPerOp > max {
			max = cp.Benchmark.MSPerOp
		}
	}
	base := 0.0
	if len(checkpoints) > 0 {
		base = checkpoints[0].Benchmark.MSPerOp
	}
	for i, cp := range checkpoints {
		kind := "before"
		switch i {
		case 1:
			kind = "best"
		case 2:
			kind = "latest"
		}
		rows = append(rows, MapRow{
			Label:    cp.Label,
			MS:       cp.Benchmark.MSPerOp,
			Width:    barWidth(cp.Benchmark.MSPerOp, max, 620),
			DeltaPct: pctChange(base, cp.Benchmark.MSPerOp),
			Kind:     kind,
		})
	}
	return rows
}

func pctChange(before, after float64) float64 {
	if before == 0 {
		return 0
	}
	return ((after - before) / before) * 100
}

func barWidth(value, max, full float64) int {
	if max <= 0 {
		return 0
	}
	w := int(math.Round((value / max) * full))
	if value > 0 && w < 20 {
		w = 20
	}
	if w > int(full) {
		w = int(full)
	}
	return w
}

func fmtMS(v float64) string {
	if v >= 100 {
		return fmt.Sprintf("%.1f ms", v)
	}
	return fmt.Sprintf("%.2f ms", v)
}

func fmtPct(v float64) string {
	if v > 0 {
		return fmt.Sprintf("+%.1f%%", v)
	}
	return fmt.Sprintf("%.1f%%", v)
}

func deltaClass(v float64) string {
	if v > 0 {
		return "regress"
	}
	return "improve"
}

func addY(base, index, step int) int {
	return base + (index * step)
}

func mapLabelY(index int) int {
	return 92 + (index * 42)
}

func mapTrackY(index int) int {
	return 78 + (index * 42)
}

func mapTextY(index int) int {
	return 95 + (index * 42)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

const svgTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="940" viewBox="0 0 1200 940" role="img" aria-labelledby="title desc">
  <title id="title">{{.Title}}</title>
  <desc id="desc">Benchmark summary for Joker interpreter optimization work, including detailed before, best-after, and latest values for the map update benchmark. Supports light and dark mode.</desc>
  <style>
    :root {
      color-scheme: light dark;
      --bg: #f6f8fc;
      --panel: #ffffff;
      --row: #f0f4fb;
      --track: #dfe7f5;
      --border: #c7d3ea;
      --text: #172033;
      --muted: #55627c;
      --baseline: #7f93b8;
      --current: #23a566;
      --before: #7f93b8;
      --best: #23a566;
      --latest: #d69a23;
    }
    @media (prefers-color-scheme: dark) {
      :root {
        --bg: #0b1020;
        --panel: #11182b;
        --row: #0f1728;
        --track: #1c2740;
        --border: #25324f;
        --text: #e3ecff;
        --muted: #9db0d6;
        --baseline: #7288b3;
        --current: #37c67e;
        --before: #7288b3;
        --best: #37c67e;
        --latest: #efb64d;
      }
    }
    svg { background: var(--bg); }
    .panel { fill: var(--panel); stroke: var(--border); stroke-width: 1; }
    .row { fill: var(--row); stroke: var(--border); stroke-width: 1; }
    .track { fill: var(--track); }
    .baseline { fill: var(--baseline); }
    .current { fill: var(--current); }
    .before { fill: var(--before); }
    .best { fill: var(--best); }
    .latest { fill: var(--latest); }
    .text { fill: var(--text); font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
    .muted { fill: var(--muted); font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
    .title { font-size: 28px; font-weight: 700; }
    .subtitle { font-size: 15px; font-weight: 500; }
    .section { font-size: 18px; font-weight: 700; }
    .metric { font-size: 16px; font-weight: 700; }
    .small { font-size: 14px; font-weight: 500; }
    .tiny { font-size: 12px; font-weight: 500; }
    .improve { fill: var(--current); font-size: 18px; font-weight: 800; }
    .regress { fill: var(--latest); font-size: 18px; font-weight: 800; }
  </style>

  <rect x="0" y="0" width="1200" height="940" fill="var(--bg)" />
  <rect class="panel" x="28" y="28" width="1144" height="884" rx="20" />

  <text class="text title" x="56" y="76">{{.Title}}</text>
  <text class="muted subtitle" x="56" y="104">{{.Subtitle}}</text>
  <text class="muted tiny" x="56" y="124">Host: {{.Host}}</text>

  <g transform="translate(56,142)">
    <rect class="baseline" x="0" y="0" width="18" height="18" rx="4" />
    <text class="text small" x="28" y="14">Baseline stable checkpoint</text>
    <rect class="current" x="252" y="0" width="18" height="18" rx="4" />
    <text class="text small" x="280" y="14">Latest checkpoint</text>
    <rect class="before" x="468" y="0" width="18" height="18" rx="4" />
    <text class="text small" x="496" y="14">Before map fastpaths</text>
    <rect class="best" x="714" y="0" width="18" height="18" rx="4" />
    <text class="text small" x="742" y="14">Best after</text>
    <rect class="latest" x="866" y="0" width="18" height="18" rx="4" />
    <text class="text small" x="894" y="14">Latest map value</text>
  </g>

  {{range $i, $row := .Rows}}
  <g transform="translate(48,{{addY 182 $i 152}})">
    <rect class="row" x="0" y="0" width="1104" height="136" rx="16" />
    <text class="text section" x="20" y="30">{{$row.Label}}</text>
    <text class="{{deltaCls $row.DeltaPct}}" x="1058" y="30" text-anchor="end">{{fmtPct $row.DeltaPct}}</text>
    <text class="muted tiny" x="20" y="58">Baseline</text>
    <rect class="track" x="120" y="44" width="620" height="22" rx="8" />
    <rect class="baseline" x="120" y="44" width="620" height="22" rx="8" />
    <text class="muted tiny" x="20" y="96">Latest</text>
    <rect class="track" x="120" y="82" width="620" height="22" rx="8" />
    <rect class="current" x="120" y="82" width="{{$row.CurrentWidth}}" height="22" rx="8" />
    <text class="text metric" x="790" y="61">{{fmtMS $row.BaselineMS}}</text><text class="muted small" x="900" y="61">baseline</text>
    <text class="text metric" x="790" y="99">{{fmtMS $row.LatestMS}}</text><text class="muted small" x="900" y="99">latest</text>
  </g>
  {{end}}

  <g transform="translate(48,646)">
    <rect class="row" x="0" y="0" width="1104" height="220" rx="16" />
    <text class="text section" x="20" y="30">Map update loop timeline</text>
    <text class="muted small" x="20" y="54">Requested view: before map fastpaths • best after • latest</text>
    {{range $i, $row := .MapRows}}
    <text class="muted tiny" x="20" y="{{mapLabelY $i}}">{{$row.Label}}</text>
    <rect class="track" x="120" y="{{mapTrackY $i}}" width="620" height="22" rx="8" />
    <rect class="{{$row.Kind}}" x="120" y="{{mapTrackY $i}}" width="{{$row.Width}}" height="22" rx="8" />
    <text class="text metric" x="790" y="{{mapTextY $i}}">{{fmtMS $row.MS}}</text>
    {{if eq $i 0}}
    <text class="muted small" x="900" y="{{mapTextY $i}}">before map fastpaths</text>
    {{else if eq $i 1}}
    <text class="muted small" x="900" y="{{mapTextY $i}}">{{fmtPct $row.DeltaPct}} vs before</text>
    {{else}}
    <text class="muted small" x="900" y="{{mapTextY $i}}">{{fmtPct $row.DeltaPct}} vs before</text>
    {{end}}
    {{end}}
    <text class="muted tiny" x="20" y="206">{{.MapNote}}</text>
  </g>
</svg>
`
