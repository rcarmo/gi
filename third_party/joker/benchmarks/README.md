# Joker benchmark artifacts

This directory stores the benchmark dataset and generated chart for the Joker optimization work.

## Files

- `benchmark-history.json` — source-of-truth dataset
- `benchmark-improvements.svg` — generated SVG chart
- `generate_svg.go` — tiny generator that reads the JSON and writes the SVG

## Regenerate the chart

From `third_party/joker/`:

```bash
go run ./benchmarks/generate_svg.go ./benchmarks
```

## Update workflow

1. Run the stable benchmark command.
2. Add/update checkpoint values in `benchmark-history.json`.
3. Regenerate the SVG with `generate_svg.go`.
4. Commit both files.
