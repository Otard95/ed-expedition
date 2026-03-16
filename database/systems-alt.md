# Galaxy Database Alternatives

This document tracks alternatives to the current custom binary format (`systems.bin` + `systems.idx` + `names.bin` + `names.trie`).

---

## Goals

- Compare implementation complexity and maintenance cost.
- Compare import/build speed for full Spansh dump.
- Compare query performance for expedition use cases.
- Track packaging/runtime impact for Wails desktop releases.

---

## Option A: SQLite (current experiment)

### Summary

Use a single SQLite DB file with table:

- `id`
- `hilbert_index`
- `name`
- `x`
- `y`
- `z`
- `star_class`

Index on `hilbert_index`.

### Pros

- Serverless, single file.
- Mature and predictable behavior.
- Easy operational model.
- Pure-Go option available (`modernc.org/sqlite`) for simpler cross-platform builds.

### Cons

- Row-store; may be slower for wide analytic scans compared to columnar engines.
- Full ingest can be slow depending on insert method and PRAGMAs.

### Notes

- Experimental loader/bench command exists: `cmd/sqlite-test/main.go`.

---

## Option B: DuckDB

### Summary

Use DuckDB as embedded OLAP engine with same logical schema.

### Pros

- Columnar storage + vectorized execution.
- Very strong scan/aggregation performance.
- Supports high-throughput bulk ingest via Appender API.

### Cons

- Go client is not pure-Go in the same sense as `modernc.org/sqlite`.
- Native/CGO packaging concerns can increase cross-platform release complexity.
- Requires explicit validation in Wails CI/release matrix.

### Notes

- Go client docs: https://duckdb.org/docs/stable/clients/go
- Go package: `github.com/duckdb/duckdb-go/v2`

---

## Option C: Keep Custom Binary Format

### Summary

Continue with current design (`systems.bin` sorted by Hilbert, sparse index, names file, trie).

### Pros

- Full control over format and access patterns.
- Minimal runtime dependencies.
- Directly optimized for target query pattern.

### Cons

- Highest implementation complexity.
- More custom correctness/performance work.
- Requires custom tooling for inspection/debugging.

---

## Evaluation Checklist

For each option, capture:

- Build/import throughput (systems/sec)
- Total build time for full dump
- DB size on disk
- Query latency for Hilbert range search (p50/p95)
- Prefix/name lookup latency (if applicable)
- Memory usage during build and query
- Cross-platform build/release friction in Wails

---

## Open Questions

- Is Hilbert range lookup selectivity high enough that row-store indexes beat columnar scans?
- Does DuckDB packaging complexity meaningfully impact release reliability for Linux/macOS/Windows?
- Is custom format complexity justified by measurable wins over SQLite/DuckDB?
