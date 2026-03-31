# Galaxy Database

SQLite-based storage for ~170 million star systems with spatial indexing and name lookup.

**Driver:** `modernc.org/sqlite` (pure-Go, no CGO)

**Location:** `DataDir/galaxy.sqlite` (see `directories.go` for OS-specific paths)

---

## Schema

```sql
CREATE TABLE systems (
    id            INTEGER PRIMARY KEY,  -- Elite Dangerous id64
    hilbert_index INTEGER NOT NULL,     -- Hilbert curve key (order 20, 60 bits)
    name          TEXT    NOT NULL,
    x             INTEGER NOT NULL,     -- normalized, scaled by 10
    y             INTEGER NOT NULL,
    z             INTEGER NOT NULL,
    star_class    INTEGER NOT NULL      -- uint8 enum (see systems.md)
);

CREATE INDEX idx_systems_hilbert ON systems(hilbert_index);
CREATE INDEX idx_systems_name    ON systems(name);
```

### Coordinate System

Coordinates are normalized to positive integers and scaled for 0.1 ly precision:

```
origin:  (-43000.0, -30000.0, -24000.0)
scale:   10 (0.1 ly precision)
formula: scaled = round((coord - origin) * 10)
```

Constants defined in `galaxy.go`: `OriginX`, `OriginY`, `OriginZ`, `CoordScale`.

### Hilbert Curve

Order 20 Hilbert curve mapping 3D coordinates to a 1D sort key (60 bits, fits in uint64). Nearby systems in 3D space have nearby Hilbert keys, enabling efficient spatial range queries via B-tree index scan.

Constants: `HilbertOrder = 20`, `HilbertBits = 60`.

Library: `gonum.org/v1/gonum/spatial/curve`.

---

## Architecture

### GalaxyDB / GalaxyTx / galaxyQuerier

Polymorphic wrapper pattern where `GalaxyDB` (wraps `*sql.DB`) and `GalaxyTx` (wraps `*sql.Tx`) share feature methods via an embedded unexported `galaxyQuerier` struct.

```
GalaxyDB                    GalaxyTx
  *sql.DB (embedded)          *sql.Tx (embedded)
  galaxyQuerier {               galaxyQuerier {
    q: queryable (*sql.DB)        q: queryable (*sql.Tx)
  }                             }
```

- **`galaxyQuerier`** holds all feature methods: schema (`EnsureSystemsTable`, `EnsureSystemsIndexes`), introspection (`ListTables`, `ListIndexesForTable`), prepared statements (`PrepareSystemInsert`), and future query methods.
- **`GalaxyDB`** adds lifecycle (`Close()`) and transaction creation (`Begin() -> *GalaxyTx`).
- **`GalaxyTx`** adds transaction lifecycle (`Commit()`, `Rollback()`) via promoted `*sql.Tx` methods.
- Both embed `*sql.DB` / `*sql.Tx` directly, so raw `Query`, `Exec`, etc. are promoted for free.

The `queryable` interface bridges `*sql.DB` and `*sql.Tx` (Go's stdlib doesn't define a shared interface for their common methods).

### Files

| File | Purpose |
|------|---------|
| `galaxy.go` | `GalaxyDB`, `GalaxyTx`, `galaxyQuerier`, `OpenGalaxyDB()`, constants |
| `galaxy_build.go` | `GalaxyBuildManager` — build pipeline |
| `galaxy_parser.go` | Streaming gzip JSON parser for Spansh dump |
| `systems.go` | `System` struct |

---

## Build Pipeline

Source: Spansh galaxy dump (`systems.json.gz`, ~5 GB compressed, ~174M systems).

### Phases

```
pending  ->  in_progress  ->  finalize  ->  done
             (insert rows)    (create indexes)
```

State persisted to `CacheDir/build.state.json`. On startup without state file, phase is probed from database (table existence + index existence).

### Pipeline Architecture

Three-stage concurrent pipeline:

```
Reader (1 goroutine)          Transform (N goroutines)       Writer (1 goroutine)
  gzip decompress               normalize coords              batched inserts
  JSON parse                     compute hilbert key           100k rows per tx
  -> rawSystemChan (50)          parse star class              explicit transactions
                                 -> systemChan (100k)
```

- Transform workers: `GOMAXPROCS/4`, clamped to `[1, 8]`.
- Writer uses explicit transactions with 100k row batches for performance.
- `PrepareSystemInsert()` uses `ON CONFLICT(id) DO NOTHING` for restart safety.

### Orchestration

Owned by `GalaxyBuildManager`. Created via `NewGalaxyBuildManager(inputPath, logger, options)`.

The service layer (`services/galaxy.go`) owns the manager lifecycle and coordinates download + build sequencing.

---

## Query Patterns (planned)

### System name autocomplete

Use `idx_systems_name` with `LIKE 'prefix%'` or similar. SQLite's B-tree index supports prefix matching efficiently.

### Spatial queries (nearest systems)

Use `idx_systems_hilbert` for range scans around a target Hilbert key, then filter by exact 3D Euclidean distance. The Hilbert curve's locality preservation means nearby 3D points cluster in the index.
