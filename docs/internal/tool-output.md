# Tool output

Gi bounds tool text sent to the model and stored as tool results to 50 KiB and
2,000 lines, retaining the head. Cuts preserve whole lines where possible and
never split a UTF-8 scalar. Invalid UTF-8 is replaced before bounding. The same
policy applies to successful results, tool errors and hook-supplied or replaced
results. Image blocks keep their separate attachment limits.

The result payload's `output_policy` records original/retained byte and line
counts, limits, retention direction and whether truncation occurred. Gi does
not append truncation prose to the tool's text. Tool-owned `details`, including
spill-file references, remain separate from harness diagnostics.

Browser progress uses a separate tail preview (100 lines, 12 KiB). Writes are
serial, persist before notification and coalesce at a size-dependent interval
with a 100 ms minimum. A terminal result replaces interim progress, including
an empty result or a result replaced by a hook. Tool completion joins pending
progress writes. A failed write remains an error at final flush.
