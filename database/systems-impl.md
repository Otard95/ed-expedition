# Systems Database - Implementation

Implementation details for building the systems database. See `systems.md` for the format specification.

---

## Build Pipeline Overview

Three phases, each independently resumable:

```
Phase 1: Download     Phase 2: Process           Phase 3: Finalize
────────────────      ─────────────────          ─────────────────
HTTP (gzip)           raw.gz                     buckets/
    │                     │                      names.bin
    ▼                     ▼                          │
raw.gz                gunzip stream                  ▼ (parallel)
                          │                     ┌─────────────────┐
                          ▼                     │ sort buckets    │
                      JSON parse                │ → systems.bin   │
                          │                     │ → systems.idx   │
                          ▼                     ├─────────────────┤
                      buckets/                  │ names.bin       │
                      names.bin                 │ → names.trie    │
                                                └─────────────────┘
```

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
  systems/
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

### State File: `process.state.json`

```json
{
  "last_system_id64": 5306398479282,
  "names_bin_size": 1500000000
}
```

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
5. On completion: delete state file

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

## Phase 3: Finalize

Sort buckets and build final files. Two independent tasks, can run in parallel.

### Task A: Build systems.bin + systems.idx

```
for bucket in 0..999:  // can parallelize sorting
    records = read_all(bucket)
    sort(records, by=hilbert_key)

// sequential concatenation
for bucket in 0..999:
    append(systems.bin, sorted_bucket)
    if first_record:
        append(systems.idx, {hilbert_key, offset})
```

### Task B: Build names.trie

```
// Collect name → system_offset mappings
for each system in systems.bin:
    name = read_name(names.bin, system.name_offset)
    mappings.append({name, system_offset})

// Sort alphabetically
sort(mappings, by=name)

// Build trie
trie = build_trie(mappings)
write(names.trie, trie)
```

### Parallelization

- Bucket sorts are independent - parallelize up to available cores/memory
- Task A and Task B are independent - can run concurrently
- Trie build needs all names in memory (~8-16 GB RAM)

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
