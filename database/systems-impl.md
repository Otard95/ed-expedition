# Galaxy Database - Implementation

Implementation details for building the galaxy database. See `systems.md` for the format specification.

---

## Build Pipeline Overview

Five phases, each independently resumable:

```
Phase 1: Download     Phase 2: Process     Phase 3: Compile     Phase 4: Index      Phase 5: Finalize
────────────────      ─────────────────    ────────────────     ──────────────      ─────────────────
HTTP (gzip)           raw.gz               buckets/             systems.bin         cache/*
    │                     │                    │                + names.bin (mmap)      │
    ▼                     ▼                    ▼                    │                    ▼
raw.gz                gunzip stream        sort buckets             ▼               move to data/
                          │                → systems.bin        names.trie          delete state
                          ▼                → systems.idx
                      JSON parse
                          │
                          ▼
                      buckets/
                      names.bin
```

State flow: `pending → process → compile → index → finalize → done`

---

## Directory Structure

```
~/.cache/ed-expedition/          # Intermediate files (deletable)
  raw.gz                         # Downloaded source
  raw.gz.<hash>.partial          # Incomplete download (hash = hash(etag+url))
  buckets/
    00/
      000.bin ... 049.bin
    01/
      050.bin ... 099.bin
    ...
    19/
      950.bin ... 999.bin
  names.bin
  process.state.json

~/.local/share/ed-expedition/    # Final database
  galaxy/
    systems.bin
    systems.idx
    names.bin
    names.trie
```

---

## Phase 1: Download

Download the gzipped source file with HTTP Range resume.

### Input
- Source URL (Spansh galaxy dump)

### Output
- `raw.gz`

### Resume Mechanism

Stateless resume using filename encoding:

```
Partial file: raw.gz.<hash>.partial
Hash: hash(etag + url)
Cursor: file size
```

### Algorithm

1. HEAD request to get `ETag`, verify `Accept-Ranges: bytes`
2. Compute `hash(etag + url)`
3. If `raw.gz.{hash}.partial` exists:
   - Resume via `Range: bytes={file_size}-`
   - Append to partial file
4. Else:
   - Create new `raw.gz.{hash}.partial`
5. On completion: rename to `raw.gz`

> **TODO:** Clean up stale `.partial` files from previous attempts with different etags. Low priority - cache dir is deleted after successful build anyway.

---

## Phase 2: Process

Stream decompress and transform into intermediate files.

### Input
- `raw.gz`

### Output
- `buckets/{group}/{bucket}.bin` - unsorted system records (33 bytes each)
- `names.bin` - raw system names, null-terminated

### Bucket Layout

1000 buckets across 20 groups (50 per group):

```go
func bucketIndex(hilbertKey uint64) int {
    const numBuckets = 1000
    const maxKey = 1 << 60  // order 20
    return int(hilbertKey / (maxKey / numBuckets))
}

func bucketPath(index int) string {
    return fmt.Sprintf("buckets/%02d/%03d.bin", index/50, index)
}
```

### State File: `build.state.json`

Persisted throughout the entire build lifecycle. Single source of truth for progress.

**Phase 2 (process):**
```json
{
  "phase": "process",
  "last_system_id64": 5306398479282,
  "names_bin_size": 1500000000
}
```

**Phase 3 (compile):**
```json
{
  "phase": "compile",
  "sorted_buckets": [0, 1, 2, 154],
  "systems_bin_complete": false
}
```

**Phase 4 (index):**
```json
{
  "phase": "index"
}
```

**Phase 5 (finalize):**
```json
{
  "phase": "finalize"
}
```

**Completed:**
```json
{
  "phase": "done"
}
```

### State Resolution (on startup)

1. `build.state.json` exists → resume from recorded phase
2. No state file, `systems.bin` exists in data dir → done
3. Neither → pending

### Algorithm

1. Open `raw.gz`, wrap in gzip reader
2. Parse JSON by counting braces (don't assume line-delimited format)
3. If resuming:
   - Iterate systems until `id64 == last_system_id64`, then continue from next
   - Truncate `names.bin` to `names_bin_size`
   - Validate bucket files, truncate if needed (see below)
4. For each system:
   - Parse (id, name, coords, star_class)
   - Normalize coordinates, compute Hilbert key
   - Append name to `names.bin`, record offset
   - Append 33-byte record to bucket file
   - Periodically update state file
5. On completion: transition state to `compile` phase

### Bucket File Validation

On resume, validate each bucket file:

```go
func validateBucket(path string) error {
    size := fileSize(path)
    if size % 33 != 0 {
        // Truncate to last complete record
        truncate(path, (size / 33) * 33)
    }
    return nil
}
```

Partial trailing records are from crashes - the system will be re-processed.

### JSON Parsing

Don't assume line-delimited format. Parse by tracking brace depth:

```go
depth := 0
for each byte:
    if '{': depth++
    if '}': depth--
    if depth == 0 && just_closed:
        // Complete object, parse it
```

Handle strings properly (braces inside strings don't count).

---

## Phase 3: Compile

Sort buckets and build `systems.bin` + `systems.idx`.

### Task A: Sort Buckets (resumable)

```
for bucket in 0..999:
    if bucket in sorted_buckets: skip
    records = read_all(bucket)
    sort(records, by=hilbert_key)
    write_back(bucket, records)
    sorted_buckets.append(bucket)
    update state file
```

Sorting is in-place (read bucket, sort, write back). Each completed bucket is recorded in `sorted_buckets` for resume. Can parallelize across cores.

### Task B: Build systems.bin + systems.idx (resumable)

Requires all buckets sorted (Task A complete).

```
for bucket in 0..999:
    append(systems.bin, sorted_bucket)
    if first_record_in_bucket:
        append(systems.idx, {hilbert_key, offset})

systems_bin_complete = true
update state file
```

On resume: if `systems_bin_complete` is false, delete partial `systems.bin` and `systems.idx`, re-concatenate from sorted buckets. Sequential concat is fast enough that partial resume isn't worth the complexity.

### Completion

When `systems_bin_complete` is true:
- Transition state to `index` phase

> **Note:** Consider deferring all file moves (names.bin, systems.bin, systems.idx) to a final cleanup step after Phase 4 completes. This would make the cache → data dir transition atomic and easier to reason about.

### Parallelization

- Bucket sorts (Task A) are independent — parallelize up to available cores/memory
- Task B depends on Task A completing

---

## Phase 4: Index

Build `names.trie` from `systems.bin` and `names.bin`.

### Algorithm (all-or-nothing)

```
mmap names.bin as names_data
for each system in systems.bin:
    name = read_null_terminated(names_data[system.name_offset:])
    mappings.append({name, system_offset})

sort(mappings, by=name)
trie = build_trie(mappings)
write(names.trie, trie)
```

`names.bin` is memory-mapped rather than loaded into a single allocation. This avoids potential issues with allocating ~3 GB of contiguous memory. The OS handles paging — only accessed portions are loaded into RAM.

Not resumable — must complete in one pass. On crash, restart from scratch. Requires ~8-16 GB RAM for the in-memory sort of mappings.

### Completion

When `names.trie` is written:
- Transition state to `finalize` phase

---

## Phase 5: Finalize

Move all output files from cache to data dir, replacing any existing database.

### Algorithm

```
move cache/systems.bin  → data/galaxy/systems.bin
move cache/systems.idx  → data/galaxy/systems.idx
move cache/names.bin    → data/galaxy/names.bin
move cache/names.trie   → data/galaxy/names.trie
delete build.state.json
```

The existing database remains fully functional until this phase completes. If interrupted, re-running finalize will complete the moves.

### Completion

When all files are moved:
- Delete `build.state.json`
- State becomes `done` (inferred from presence of `systems.bin` in data dir and absence of state file)

---

## Resource Requirements

| Resource | Minimum |
|----------|---------|
| Disk (during build) | ~15 GB |
| Disk (final) | ~10 GB |
| RAM (trie build) | ~8 GB |
| Network | ~5.2 GB download |

Present requirements to user before starting. Fail fast if not met.

---

## Cleanup

After successful build:
- Delete `~/.cache/ed-expedition/` entirely
- Or keep `raw.gz` for potential rebuilds

---

## Future Optimizations

1. Parallel bucket sorting
2. Streaming trie build (reduce RAM requirement)
3. Incremental updates (detect changed systems)
4. Compressed storage
