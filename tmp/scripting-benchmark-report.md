# Scripting benchmark suite

This report compares **Go**, **Python**, **Goja**, and **Joker** on a small set of common scripting-style workloads.

Methodology:
- 3 runs per runtime/task
- median reported as the main comparison metric
- each runtime executes the full task in a single invocation
- Goja/Joker use gi's embedded scripting runners
- Python uses `python3 -c ...`

## Summary

Across these tasks, the rough ranking is consistently **Go > Python > Goja > Joker**. Python is often an order of magnitude slower than Go but still practical; Goja is generally slower than Python for these workloads; Joker is far slower for CPU-heavy tasks.

## recursive-fib

CPU-bound recursion: fib(28) repeated 5 times inside one invocation.

| Runtime | Result | Median | Average | Runs | Slowdown vs Go |
|---|---:|---:|---:|---|---:|
| go | 1589055 | 10.038843ms | 9.557565ms | 8.13592ms, 10.038843ms, 10.497933ms | 1.0x |
| python | 1589055 | 218.953464ms | 225.953118ms | 218.953464ms, 247.524954ms, 211.380937ms | 21.8x |
| goja | 1589055 | 977.391092ms | 958.857002ms | 1.117410473s, 977.391092ms, 781.769442ms | 97.4x |
| joker | 1589055 | 6.240598032s | 6.209546711s | 6.00688258s, 6.381159523s, 6.240598032s | 621.6x |

## arith-loop

Tight arithmetic loop: 1,000,000 iterations accumulating a modular sum.

| Runtime | Result | Median | Average | Runs | Slowdown vs Go |
|---|---:|---:|---:|---|---:|
| go | 4999995 | 1.001976ms | 1.027399ms | 1.001976ms, 1.108348ms, 971.874µs | 1.0x |
| python | 4999995 | 99.820718ms | 108.327837ms | 133.168908ms, 99.820718ms, 91.993887ms | 99.6x |
| goja | 4999995 | 324.07924ms | 325.383892ms | 296.042012ms, 324.07924ms, 356.030424ms | 323.4x |
| joker | 4999995 | 4.178578315s | 4.14677771s | 4.054668232s, 4.178578315s, 4.207086583s | 4170.3x |

## json-parse-sum

JSON workload: parse a 500-record JSON array 20 times and sum the "value" fields.

| Runtime | Result | Median | Average | Runs | Slowdown vs Go |
|---|---:|---:|---:|---|---:|
| go | 89300 | 4.484093ms | 4.486331ms | 4.484093ms, 4.4725ms, 4.502401ms | 1.0x |
| python | 89300 | 17.595529ms | 17.505103ms | 17.595529ms, 17.607115ms, 17.312665ms | 3.9x |
| goja | 89300 | 39.524466ms | 38.860944ms | 41.063563ms, 39.524466ms, 35.994803ms | 8.8x |
| joker | 89300 | 29.731897ms | 27.471018ms | 29.731897ms, 22.488744ms, 30.192414ms | 6.6x |

## word-frequency

Text processing: split a 12k-word string and build a frequency map.

| Runtime | Result | Median | Average | Runs | Slowdown vs Go |
|---|---:|---:|---:|---|---:|
| go | 3000 | 431.018µs | 440.207µs | 431.018µs, 416.886µs, 472.717µs | 1.0x |
| python | 3000 | 10.828984ms | 10.834937ms | 11.276855ms, 10.398972ms, 10.828984ms | 25.1x |
| goja | 3000 | 12.904202ms | 12.302523ms | 10.839939ms, 12.904202ms, 13.163429ms | 29.9x |
| joker | 3000 | 2.409329972s | 2.42158177s | 2.486236968s, 2.369178371s, 2.409329972s | 5589.9x |

## Interpretation

- **Go** is the clear performance baseline and best fit for compute-heavy or latency-sensitive logic.
- **Python** is much slower than Go, but still reasonable for many orchestration and data-processing tasks.
- **Goja** is viable for embedded scripting and control logic, but not competitive with Python on these heavier microbenchmarks.
- **Joker** is expressive and useful for scripting ergonomics, but should not be treated as a high-performance compute runtime.
