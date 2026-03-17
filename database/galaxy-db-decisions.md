# Galaxy Database - Decisions

Design decisions for the SQLite galaxy database. See `galaxy-db.md` for the spec.

For decisions about the earlier custom binary format approach, see `systems-decisions.md`.

---

## SQLite over Alternatives

Three approaches were evaluated for storing ~170M star systems:

### Option A: SQLite (chosen)

Single SQLite DB file with `modernc.org/sqlite` (pure-Go, no CGO).

**Pros:**
- Single-file database with mature tooling and predictable behavior
- No custom file format correctness burden (trie, sparse index, bucket sort)
- No ~8 GB RAM requirement for trie construction
- B-tree indexes cover both spatial queries (Hilbert) and name lookup
- Pure-Go driver avoids CGO cross-compilation complexity in Wails CI

**Cons:**
- Row-store; slower for wide analytic scans than columnar engines
- Full ingest can be slow depending on insert method and PRAGMAs

### Option B: DuckDB

Embedded OLAP engine with columnar storage and vectorized execution.

**Rejected because:**
- Go client requires CGO, increasing cross-platform release complexity
- Needs explicit validation in Wails CI/release matrix for Linux/macOS/Windows
- Columnar strengths (scan/aggregation) don't match our workload (point lookups, narrow range scans)

### Option C: Custom binary format

Purpose-built files (`systems.bin` sorted by Hilbert, sparse index, `names.trie`, `names.bin`). See `systems.md`, `systems-decisions.md`, `systems-impl.md` for the full design.

**Rejected because:**
- Highest implementation complexity (bucket sort, trie construction, custom binary serialization)
- Requires ~8 GB RAM for in-memory trie build
- Custom tooling needed for inspection and debugging
- Correctness burden not justified given SQLite covers the query patterns

### Verdict

Expedition use cases are point lookups and narrow range scans — exactly what B-trees excel at. SQLite's row-store trade-off is irrelevant for this workload.

---

## GalaxyDB Polymorphic Pattern

Both `*sql.DB` and `*sql.Tx` share identical method signatures (`Prepare`, `Exec`, `Query`, `QueryRow`) but Go's stdlib doesn't define a shared interface.

**Chosen:** Define a `queryable` interface and embed it in `galaxyQuerier`. Both `GalaxyDB` and `GalaxyTx` embed `galaxyQuerier` with their respective backing type. Feature methods live on `galaxyQuerier` — the split between `GalaxyDB` and `GalaxyTx` is only about lifecycle.

**Why not methods directly on GalaxyDB/GalaxyTx:** Would duplicate every feature method.

**Why embedded `*sql.DB`/`*sql.Tx` doesn't collide:** Interface fields (`q queryable`) don't promote their methods — only embedded types do. So `GalaxyDB`'s promoted `Query` comes from `*sql.DB`, not from `galaxyQuerier.q`.

---

## Transaction Batching for Build

SQLite implicit transactions = 1 fsync per INSERT. For 174M rows, that's catastrophic.

**Chosen:** Explicit transactions with 100k row batches. Writer commits every 100k rows, then begins a new transaction and re-prepares the insert statement.

A `db.Prepare()`'d statement used inside a `tx` gets silently re-prepared by Go's `database/sql` — so preparing on the tx directly is the correct approach.

---

## Index Creation: After Inserts, Not Before

**Tested 2026-03-17** with 1M system mock dataset, 4 transform workers.

### Index-after (current approach)

| Phase | Duration |
|-------|----------|
| Insert 1M rows | ~6.8s |
| Create indexes | ~4.7s |
| **Total** | **~11.4-11.5s** |

Two runs: 11.42s, 11.51s. Very consistent.

### Index-before (indexes exist during inserts)

| Phase | Duration |
|-------|----------|
| Insert 1M rows (with indexes maintained) | ~24.0-24.8s |
| Create indexes | n/a |
| **Total** | **~24.0-24.8s** |

Two runs: 24.06s, 24.82s.

### Verdict

Index-before is **~2.1x slower**. Each INSERT must update two B-trees (hilbert + name) with essentially random keys, thrashing the page cache. The simplification benefits (single progress dimension, fewer build phases) don't justify doubling build time.

### Useful ratio for progress estimation

The insert-to-index time ratio is stable at roughly **59% insert / 41% index** (6.8s / 4.7s). This can be used to synthesize a progress bar during the finalize phase:

```
estimated_index_duration = observed_insert_duration * 0.69
```

This ratio should be re-validated if the schema, index count, or dataset characteristics change significantly, but for the current 2-index setup on ~170M rows it provides a reasonable approximation.
