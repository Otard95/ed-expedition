# Galaxy Database - Custom Binary Format (Archived)

> **Status: Not implemented.** SQLite was chosen instead. See `galaxy-db.md` for the current approach and `galaxy-db-decisions.md` for why.
>
> This document preserves the original custom binary format design for reference. The spatial indexing concepts (Hilbert curve, coordinate normalization) carry over to the SQLite implementation.

---

# Part 1: Specification

Binary file formats for storing and querying ~170 million star systems with spatial indexing and name lookup.

## Overview

The database consists of four files:

| File | Purpose | Size (estimated) |
|------|---------|------------------|
| `systems.bin` | Hilbert-sorted system records | ~5.5 GB |
| `systems.idx` | Sparse index for fast lookup | ~55 KB |
| `names.trie` | Prefix trie for name search/autocomplete | ~500 MB - 1 GB |
| `names.bin` | Raw system names, null-terminated | ~2.5 - 3.5 GB |

**Total: ~9-10 GB**

---

## File Header

All binary files share a common 8-byte header for identification and versioning.

| Field | Type | Bytes | Description |
|-------|------|-------|-------------|
| magic | [4]u8 | 4 | `"EDEX"` (0x45 0x44 0x45 0x58) |
| file_type | [2]u8 | 2 | File type identifier (see below) |
| version | u16 | 2 | Format version, little-endian |

**File type identifiers:**
- `"SB"` (0x53 0x42) — systems.bin
- `"SI"` (0x53 0x49) — systems.idx
- `"NB"` (0x4E 0x42) — names.bin
- `"NT"` (0x4E 0x54) — names.trie

**Version:** Increment when the format changes incompatibly. Readers should reject unknown versions.

**Example:** systems.bin version 1 header in hex: `45 44 45 58 53 42 01 00`

> **Note:** All byte offsets in this spec (e.g., `name_offset` in system records, offsets in `systems.idx`) are relative to data start (byte 8), not file start.

---

## Coordinate System

All coordinates are normalized and scaled for Hilbert curve compatibility.

**Source data:**
- Origin: `(-42213.8, -29359.8, -23405.0)` ly
- Extent: `(82718, 68878, 89035)` ly per axis
- Precision: 0.03125 ly (1/32 ly)

**Normalization:**
- Shift all coordinates to positive: `normalized = coord - origin`
- Scale by 10 for 0.1 ly precision: `scaled = round(normalized * 10)`
- Result: unsigned integers suitable for Hilbert curve

**Hilbert curve parameters:**
- Precision: 0.1 ly (scale factor 10)
- Max cells per axis: ~900,000
- Order: 20 (2^20 = 1,048,576 cells per axis)
- Index bits: 60 (fits in u64)

> **Implementation note:** We use `gonum.org/v1/gonum/spatial/curve` for Hilbert indexing. Its `Hilbert3D.Len()` return type is `int` and overflows at order >= 21 on 64-bit architectures (>= 11 on 32-bit). Order 20 therefore requires a 64-bit build target.

> **Note:** Multiple systems may share the same Hilbert index due to 0.1 ly quantization (minimum system distance is 0.03125 ly). This is fine - Hilbert key is a sort key, not a unique identifier.

> **Tuning:** Order 21 (0.05 ly precision) uses 63 bits and reduces collisions. Order 19 (0.2 ly precision) uses 57 bits if smaller keys are needed.

---

## systems.bin

Fixed-size records sorted by Hilbert key. Enables spatial queries via binary search + range scan.

**File structure:** 8-byte header (`EDEXSB`) + N x 37-byte records.

### Record Format (37 bytes)

| Field | Type | Bytes | Description |
|-------|------|-------|-------------|
| hilbert_key | u64 | 8 | Hilbert curve index (order 20, 60 bits used) |
| x | u32 | 4 | Normalized X coordinate, scaled by 10 |
| y | u32 | 4 | Normalized Y coordinate, scaled by 10 |
| z | u32 | 4 | Normalized Z coordinate, scaled by 10 |
| id | u64 | 8 | System id64 (Elite Dangerous unique identifier) |
| star_class | u8 | 1 | Star classification enum (see below) |
| name_offset | u64 | 8 | Byte offset into names.bin |

**Size:** ~170,000,000 systems x 37 bytes = **~6.29 GB**

> **Tuning:** For power-of-2 alignment (potential memory-mapped performance benefit), pad to 40 bytes (+1.19 GB). Likely not worth it.

### Star Classification Enum

Values derived from Elite Dangerous galaxy data. The enum groups related types with room for expansion.

```
0x00 = Unknown

Main Sequence (0x01 - 0x0F)
0x01 = O     "O (Blue-White) Star"
0x02 = B     "B (Blue-White) Star"
0x03 = A     "A (Blue-White) Star"
0x04 = F     "F (White) Star"
0x05 = G     "G (White-Yellow) Star"
0x06 = K     "K (Yellow-Orange) Star"
0x07 = M     "M (Red dwarf) Star"

Giants & Supergiants (0x10 - 0x1F)
0x10 = K_G   "K (Yellow-Orange giant) Star"
0x11 = M_G   "M (Red giant) Star"
0x12 = M_SG  "M (Red super giant) Star"
0x13 = A_SG  "A (Blue-White super giant) Star"
0x14 = B_SG  "B (Blue-White super giant) Star"
0x15 = F_SG  "F (White super giant) Star"
0x16 = G_SG  "G (White-Yellow super giant) Star"

Brown Dwarfs (0x20 - 0x2F)
0x20 = L     "L (Brown dwarf) Star"
0x21 = T     "T (Brown dwarf) Star"
0x22 = Y     "Y (Brown dwarf) Star"

Carbon Stars (0x30 - 0x3F)
0x30 = C     "C Star"
0x31 = CN    "CN Star"
0x32 = CJ    "CJ Star"
0x33 = MS    "MS-type Star"
0x34 = S     "S-type Star"

White Dwarfs (0x40 - 0x5F)
0x40 = D     "White Dwarf (D) Star"
0x41 = DA    "White Dwarf (DA) Star"
0x42 = DAB   "White Dwarf (DAB) Star"
0x43 = DAV   "White Dwarf (DAV) Star"
0x44 = DAZ   "White Dwarf (DAZ) Star"
0x45 = DB    "White Dwarf (DB) Star"
0x46 = DBV   "White Dwarf (DBV) Star"
0x47 = DBZ   "White Dwarf (DBZ) Star"
0x48 = DC    "White Dwarf (DC) Star"
0x49 = DCV   "White Dwarf (DCV) Star"
0x4A = DQ    "White Dwarf (DQ) Star"

Wolf-Rayet (0x60 - 0x6F)
0x60 = W     "Wolf-Rayet Star"
0x61 = WC    "Wolf-Rayet C Star"
0x62 = WN    "Wolf-Rayet N Star"
0x63 = WNC   "Wolf-Rayet NC Star"
0x64 = WO    "Wolf-Rayet O Star"

Proto Stars (0x70 - 0x7F)
0x70 = TTS   "T Tauri Star"
0x71 = AeBe  "Herbig Ae/Be Star"

Compact Objects (0x80 - 0x8F)
0x80 = N     "Neutron Star"
0x81 = BH    "Black Hole"
0x82 = SMBH  "Supermassive Black Hole"
```

### Spatial Query Algorithm

To find systems within radius R of point P:

1. Compute Hilbert key for P
2. Find bucket in `systems.idx` (linear scan or binary search, 3480 entries)
3. Lerp within bucket to estimate file offset
4. Binary search to find exact position
5. Scan outward (both directions) collecting systems
6. Filter by exact 3D distance (Hilbert locality isn't perfect at fold boundaries)
7. Stop when Hilbert keys are far enough that no more matches are possible

> **Tuning:** Scan range depends on query radius and local density. For typical 50 ly jump range queries, scanning +/-10,000 systems from the estimated position should be more than sufficient.

---

## systems.idx

Sparse index mapping Hilbert keys to file offsets. Loaded into memory at startup.

**File structure:** 8-byte header (`EDEXSI`) + N x 16-byte entries.

### Format

Array of entries, one per bucket (every 50,000 systems):

| Field | Type | Bytes | Description |
|-------|------|-------|-------------|
| hilbert_key | u64 | 8 | Hilbert key of first system in bucket |
| file_offset | u64 | 8 | Byte offset into systems.bin |

**Entry count:** ceil(~170,000,000 / 50,000) = **~3,400 entries**

**Size:** ~3,400 x 16 bytes = **~54.4 KB**

### Lookup Algorithm

```
1. Linear scan (or binary search) to find bucket containing target Hilbert key
2. Get bucket start/end keys and file offsets
3. Lerp to estimate position within bucket:
   ratio = (targetKey - startKey) / (endKey - startKey)
   estimatedOffset = startOffset + ratio * (endOffset - startOffset)
4. Binary search from estimated position
```

> **Tuning:** Bucket size trades index size vs search range.
> - 50,000 systems/bucket: ~55 KB index, ~16 binary search iterations within bucket
> - 10,000 systems/bucket: ~272 KB index, ~14 iterations
> - 100,000 systems/bucket: ~27 KB index, ~17 iterations
>
> 50,000 is a reasonable default. The lerp step makes iteration count less important.

> **Note:** The binary search step (4) can likely be optimized or eliminated:
> - We don't need exact position — we're going to scan outward anyway. Getting "close enough" (within a few thousand records) is sufficient.
> - Interpolation search may outperform binary search here since Hilbert keys are roughly uniformly distributed.
> - The lerp estimate may already be close enough to skip directly to scanning, especially for larger query radii.

---

## names.trie

Prefix trie for name lookup and autocomplete. Optimized for minimal seeks during lookup.

**File structure:** 8-byte header (`EDEXNT`) + trie data.

### Layout Principle

Siblings are stored contiguously. This means:
- Scanning siblings = sequential read (fast)
- Descending to children = single seek (unavoidable)

Lookup requires O(depth) seeks (~3-5 for typical system names), not O(siblings) seeks.

### Sibling Group Format

Each group of siblings is prefixed with a count:

| Field | Type | Bytes | Description |
|-------|------|-------|-------------|
| sibling_count | u32 | 4 | Number of nodes in this sibling group |
| nodes | [Node] | varies | `sibling_count` nodes, contiguous |

### Node Format

| Field | Type | Bytes | Description |
|-------|------|-------|-------------|
| str_len | u8 | 1 | Length of segment string (max 255, segments are much shorter) |
| segment | [u8] | str_len | The name segment (e.g., "Lyruewry", "AA-A", "h0") |
| flags | u8 | 1 | Bit 0: is_terminal (rest reserved) |
| ref | u64 | 8 | If terminal: byte offset into systems.bin. If non-terminal: byte offset to first child's sibling group |

**Node size:** 10 + str_len bytes

### Structure

- Root sibling group starts at data offset 0 (file offset 8, after header)
- Siblings within a group are sorted alphabetically
- Non-terminal nodes point to their children's sibling group
- Terminal nodes point to system records in `systems.bin`

### Example

For systems "Lyruewry AA-A h0", "Lyruewry AA-A h1", "Lyruewry BA-A h0", "Other Area XY-Z a0":

```
Offset 0 (root sibling group):
  <sibling_count: 2>
  [Node: "Lyruewry", non-terminal, children_offset=100]
  [Node: "Other Area", non-terminal, children_offset=200]

Offset 100 (Lyruewry's children):
  <sibling_count: 2>
  [Node: "AA-A", non-terminal, children_offset=300]
  [Node: "BA-A", non-terminal, children_offset=400]

Offset 200 (Other Area's children):
  <sibling_count: 1>
  [Node: "XY-Z", non-terminal, children_offset=500]

Offset 300 (AA-A's children):
  <sibling_count: 2>
  [Node: "h0", terminal, system_offset=<offset1>]
  [Node: "h1", terminal, system_offset=<offset2>]

Offset 400 (BA-A's children):
  <sibling_count: 1>
  [Node: "h0", terminal, system_offset=<offset3>]

Offset 500 (XY-Z's children):
  <sibling_count: 1>
  [Node: "a0", terminal, system_offset=<offset4>]
```

### Traversal

**Name lookup (e.g., "Lyruewry AA-A h1"):**
1. Seek to offset 0, read sibling_count
2. Sequential scan: read nodes until "Lyruewry" matches (or count exhausted)
3. Match found, non-terminal -> seek to children_offset (100)
4. Read sibling_count, sequential scan until "AA-A" matches
5. Match found, non-terminal -> seek to children_offset (300)
6. Read sibling_count, sequential scan until "h1" matches
7. Match found, terminal -> return system_offset

**Total seeks:** 3 (one per level)

**Autocomplete (e.g., prefix "Lyruewry A"):**
1. Navigate to deepest fully-matched node (as above)
2. Return first N siblings at that level matching the partial segment (already sorted alphabetically)

> **Future:** DFS to collect terminal descendants across levels if deeper autocomplete is needed.

**Name validation:** Same as lookup - system exists if lookup reaches a terminal node.

> **Note:** Reconstructing full names during autocomplete requires tracking the path from root. Keep a stack of segments during traversal.

---

## names.bin

Raw system names, null-terminated. Referenced by `name_offset` in `systems.bin`.

**File structure:** 8-byte header (`EDEXNB`) + null-terminated name strings.

### Format

```
<name1>\0<name2>\0<name3>\0...
```

No structure, no length prefixes. Each name is null-terminated. `name_offset` points to the first character of the name.

**Size estimate:** ~170,000,000 systems x ~18 chars average = **~3.1 GB**

> **Note:** Ordering doesn't matter. Names can be written in any order during build - each system record stores its specific offset.

> **Note:** This duplicates name data that's also in `names.trie`. The ~3 GB cost is accepted for simplicity - reconstructing names from the trie would require parent pointers or path tracking, adding complexity, and crucially too much IO/seek in `names.trie`.

---

## Query Patterns

### Find system by exact name

```
1. Traverse names.trie to terminal node
2. Read system_offset from terminal node
3. Seek to system_offset in systems.bin
4. Read 37-byte record
```

### Find system by id64

Not directly supported. Options:
- Full scan of systems.bin (slow, ~5.6 GB read)
- Build separate id64 index (adds ~2.7 GB for sorted id64 -> offset mapping)

> **Recommendation:** If id64 lookup is needed, add `ids.idx` mapping sorted id64 to file offset.

### Autocomplete by name prefix

```
1. Traverse names.trie to deepest fully-matched node
2. Return first N siblings at that level matching the partial segment
3. For each match: reconstruct full name from path stack
```

### Find systems near point (spatial query)

```
1. Normalize and scale query point
2. Compute Hilbert key
3. Use systems.idx to find bucket
4. Lerp + binary search to find position in systems.bin
5. Scan outward, collecting systems
6. Filter by exact 3D distance
7. Return matching systems
```

### Find systems in region (bounding box)

```
1. Compute Hilbert keys for box corners (8 keys)
2. Query range spans min to max Hilbert key
3. Scan that range in systems.bin
4. Filter by exact 3D bounds
```

> **Note:** Large regions may span most of the Hilbert curve due to fold structure. Consider subdividing the query.

---

## Partial Galaxy Support

The flat Hilbert file supports partial galaxy storage naturally:

- Only include systems in region of interest during build
- Hilbert ordering still works with gaps
- Binary search still works
- Spatial queries still work

**Trade-off:** No dynamic updates. To add/remove regions, rebuild the entire database.

---

## Build Process

> **Note:** This is the spec-level conceptual overview (4 phases). The implementation expanded this into a 5-phase pipeline with different breakdown — see Part 3 for details.

### Phase 1: Parse and Sort

1. Stream source JSON, extract: id, name, coords, star_class
2. Normalize coordinates (shift + scale)
3. Compute Hilbert key for each system
4. Write to temporary file: `[hilbert_key, x, y, z, id, star_class, name]`
5. External sort by hilbert_key (file too large for memory sort)

### Phase 2: Write systems.bin and names.bin

1. Stream sorted systems
2. For each system:
   - Write name to names.bin, record offset
   - Write record to systems.bin with name_offset
3. Every 50,000 systems: record (hilbert_key, file_offset) for index

### Phase 3: Write systems.idx

1. Write collected index entries

### Phase 4: Build names.trie

1. Stream systems.bin (or use names from Phase 1)
2. Build trie in memory (may require ~8-16 GB RAM)
3. Serialize to disk

> **Alternative for Phase 4:** Build trie incrementally during Phase 2 if names are pre-sorted alphabetically. This avoids memory pressure but requires a separate sort pass.

---

# Part 2: Design Decisions

## Spatial Indexing: Hilbert Curve

### Why Hilbert over other approaches?

**Considered:**
- **Naive 3D grid with HashMap** - Simple, but poor locality for range queries
- **Z-order curve (Morton)** - Simpler math, but worse locality preservation than Hilbert
- **Octree/KD-tree** - Good for dynamic data, but complex and pointer-heavy
- **R-tree** - Optimized for rectangles, overkill for point data

**Chosen: Hilbert curve** because:
- Best locality preservation of any space-filling curve
- Maps 3D -> 1D while keeping nearby points mostly nearby
- Enables simple sorted binary file with binary search
- No pointers, no tree structure, minimal overhead

### Why not a grid system?

We considered partitioning the galaxy into large cells with local Hilbert curves per cell. This would enable:
- 32-bit Hilbert keys (order 10-11 per cell)
- Partial galaxy storage (only download/store needed regions)

**Decision:** Start with flat Hilbert file for v1. Reasons:
- Simpler single-file design
- 64-bit keys work fine at 0.1 ly precision
- Partial galaxy can still be achieved by filtering during build
- Can add grid partitioning later if needed

### Hilbert order and precision trade-off

**Constraint:** 64-bit keys, 3D space -> max 21 bits per axis (63 bits total)

| Order | Bits | Precision | Fits in |
|-------|------|-----------|---------|
| 20 | 60 | 0.1 ly | u64 |
| 21 | 63 | 0.05 ly | u64 |
| 22 | 66 | 0.025 ly | too big |

**Chosen: Order 20 (0.1 ly precision)** because:
- Fits comfortably in u64 with room to spare
- Actual minimum system distance is 0.03125 ly, so some systems share keys
- Sharing keys is fine - Hilbert key is a sort key, not unique identifier
- Coarser precision means fewer bits, simpler math

---

## Name Storage: Trie + Flat Names File

### Why duplicate names in both trie and names.bin?

The trie stores name segments for traversal. To get a full name from a spatial query result, you'd need to walk up the trie - but tries don't have parent pointers.

**Options considered:**
1. Add parent pointers to trie - adds 8 bytes per node, complicates format
2. Store path indices in system record - fragile, assumes fixed trie depth
3. Reconstruct name during trie traversal - requires tracking path anyway
4. Separate names.bin file - simple, ~3 GB duplication accepted

**Chosen: Duplicate storage** because:
- Simplicity wins
- 3 GB is acceptable for avoiding complexity
- Each file has single responsibility: trie for lookup, names.bin for retrieval

### Why siblings-contiguous trie layout?

**Original design:** Children follow parent, siblings linked via pointers

**Problem:** Name lookup requires scanning siblings. With pointer-linked siblings, each sibling check is a seek. For a name 4 levels deep with 1000 siblings at root level, that's potentially 1000+ seeks.

**New design:** Siblings stored contiguously, sibling count prefix

**Result:** O(depth) seeks instead of O(total siblings visited). Sequential reads for sibling scanning, seeks only when descending to children.

### Why u8 for segment length?

Original spec used u16. But name segments (e.g., "Lyruewry", "AA-A", "h0") are never anywhere near 255 characters. u8 saves 1 byte per node across ~500M+ nodes.

---

## Index Design: Sparse Hilbert Index

### Why not just binary search the whole file?

You could binary search 170M records directly - that's ~27 iterations, each reading 37 bytes. On SSD, probably <1ms.

**But:** We can do better with minimal overhead.

### Why sparse index + lerp?

**Design:** Sample every 50,000 systems -> 3,400 index entries -> ~55 KB

**Lookup:**
1. Find bucket containing target key (scan 3,400 entries)
2. Lerp within bucket to estimate position
3. Binary search from estimate

**Why lerp works within buckets:** 50,000 systems in a bucket span a small, localized region of space. Density is relatively uniform at this scale, so linear interpolation gets close.

**Result:** Effectively O(1) bucket lookup + ~5-10 binary search iterations instead of ~27.

### Why not HashMap for bucket lookup?

Considered: `map[uint64]uint64` keyed by bucket's first Hilbert key.

**Problem:** You need to find which bucket *contains* a key, not exact key match. Would need to find largest key <= target. HashMap doesn't support this.

**Solution:** Sorted slice, linear scan or binary search. 3,400 entries is tiny - linear scan is nanoseconds.

---

## Build Pipeline: Three Separate Phases

### Why not stream directly from HTTP to final files?

**Original idea:** Download gzipped JSON -> decompress -> parse -> transform -> write final files, all streaming.

**Problems:**
1. `systems.bin` must be sorted by Hilbert key, but source is arbitrary order
2. Gzip streams aren't seekable - can't resume mid-stream without context
3. Coupling download and processing makes error recovery complex

### Why separate download phase?

**Key insight:** Source is gzipped. You can't resume a gzip stream mid-way without decompression context. So either:
- Download fully, then process (simple)
- Use indexed gzip format like dictzip (complex)
- Accept re-decompressing from start on resume (wasteful)

**Decision:** Download fully first. Network is the bottleneck anyway. Processing 10 GB locally takes ~10-15 minutes; downloading 5 GB takes 30-60 minutes on typical connections.

### Why bucket-based sorting?

**Problem:** 170M systems don't fit in memory for sorting.

**Options:**
1. External merge sort - complex, many temp files
2. Database (SQLite, etc.) - heavyweight dependency
3. Bucket sort with in-memory sort per bucket - simple, parallelizable

**Chosen: Bucket sort** because:
- 1000 buckets x 170k systems/bucket
- 170k x 37 bytes = ~6.3 MB per bucket - easily fits in memory
- Each bucket sort is independent - trivially parallelizable
- Final concatenation is sequential but simple

---

## Resume Strategy

### Why stateless download resume?

**Original design:** State file tracking bytes_downloaded, etag, etc.

**Insight:** All resume state can be encoded in filename:
- `raw.gz.<hash(etag+url)>.partial`
- Hash matches = same source version
- File size = bytes downloaded

**Result:** No state file needed for download phase. Simpler, fewer failure modes.

### Why track last_system_id64 for process resume?

**Considered:**
- Byte offset into decompressed stream - gzip isn't seekable
- System count - would need to count on resume, slow
- Byte offset into compressed stream - complex, gzip context issues

**Chosen: last_system_id64** because:
- Simple to track and persist
- On resume: decompress from start, skip until id64 matches, continue
- Decompression is fast, skipping is cheap
- No assumptions about ordering (use `==` not `>`)

### Why checkpoint exact bucket sizes for process resume?

**Problem:** Resume must keep three things aligned at the same logical point:
- `last_system_id64`
- `names.bin` length
- every bucket file length

If these drift, resume can duplicate systems or produce invalid `name_offset` values.

**Chosen:** Persist exact `bucket_sizes[]` and `names_bin_size` alongside `last_system_id64` at each checkpoint, then truncate to those exact sizes on resume.

**Why record-boundary truncation is not enough:** Truncating each bucket to `floor(size/37)*37` only repairs torn writes; it does not prove the bucket data corresponds to the same checkpoint as `last_system_id64`.

---

## Resource Trade-offs

### Why accept ~10 GB final size?

**Breakdown:**
- systems.bin: 170M x 37 bytes = 6.3 GB
- names.bin: 170M x ~18 chars = 3.1 GB
- names.trie: ~500 MB - 1 GB
- systems.idx: ~55 KB

**Could reduce via:**
- Compression - adds CPU overhead, complicates random access
- Delta encoding - modest savings, complexity cost
- Smaller types - already using u32 for coords, u8 for star_class

**Decision:** Accept 10 GB. Storage is cheap, simplicity is valuable. Compression can be added later if needed.

### Why require ~8 GB RAM for trie build?

Trie construction needs all names in memory (sorted). 170M names x ~50 bytes overhead = ~8 GB.

**Alternatives:**
- External sort + streaming trie build - complex, deferred to future optimization
- Memory-mapped intermediate files - still need sorting

**Decision:** Require 8 GB RAM for v1. Most dev machines have this. Document as minimum requirement.

---

## Format Choices

### Why fixed-size records in systems.bin?

Variable-length records (e.g., inline names) would require:
- Length prefixes or delimiters
- Index for random access
- Complex seeking logic

Fixed 37-byte records enable:
- Direct offset calculation: `offset = index * 37`
- Simple binary search
- Memory-mappable with predictable access patterns

### Why null-terminated names in names.bin?

**Options:**
- Length-prefixed: read length, read bytes
- Null-terminated: read until \0
- Fixed-size: pad all names to max length

**Chosen: Null-terminated** because:
- System names don't contain \0
- Slightly smaller than length-prefixed (no 2-byte prefix)
- Simple C-string compatibility
- Only accessed via known offset, never scanned

### Why u64 for name_offset?

With 170M names, growth in average name length and future dataset expansion can push names.bin beyond 4 GB.

**Decision:** Use u64 offsets now.

**Trade-off:** systems.bin grows by +4 bytes per record compared to u32 offsets.

**Why this is acceptable:** The increase is predictable and avoids future format migration pressure from offset overflow risk.

---

# Part 3: Implementation Details

## Build Pipeline Overview

Five phases, each independently resumable:

```
Phase 1: Download     Phase 2: Process     Phase 3: Compile     Phase 4: Index      Phase 5: Finalize
----------------      -----------------    ----------------     --------------      -----------------
HTTP (gzip)           raw.gz               buckets/             systems.bin         cache/*
    |                     |                    |                + names.bin (mmap)      |
    v                     v                    v                    |                    v
raw.gz                gunzip stream        sort buckets             v               move to data/
                          |                -> systems.bin        names.trie          delete state
                          v                -> systems.idx
                      JSON parse
                          |
                          v
                      buckets/
                      names.bin
```

State flow: `pending -> process -> compile -> index -> finalize -> done`

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

---

## Phase 2: Process

Stream decompress and transform into intermediate files.

### Input
- `raw.gz`

### Output
- `buckets/{group}/{bucket}.bin` - unsorted system records (37 bytes each)
- `names.bin` - raw system names, null-terminated

### Bucket Layout

1000 buckets across 20 groups (50 per group):

```go
func bucketIndex(hilbertKey uint64) int {
    const numBuckets = 1000
    maxKey := uint64(h.Len())
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
  "names_bin_size": 1500000000,
  "bucket_sizes": [0, 660, 1320]
}
```

`bucket_sizes` has one entry per bucket (1000 total), each storing the exact byte size written at checkpoint time.

**Phase 3 (compile):**
```json
{
  "phase": "compile",
  "sorted_buckets": [0, 1, 2, 154],
  "systems_bin_complete": false
}
```

**Phases 4-5 and done:**
```json
{ "phase": "index" }
{ "phase": "finalize" }
{ "phase": "done" }
```

### State Resolution (on startup)

1. `build.state.json` exists -> resume from recorded phase
2. No state file, `systems.bin` exists in data dir -> done
3. Neither -> pending

### Algorithm

1. Open `raw.gz`, wrap in gzip reader
2. Parse JSON by counting braces (don't assume line-delimited format)
3. If resuming:
   - Iterate systems until `id64 == last_system_id64`, then continue from next
   - Truncate `names.bin` to `names_bin_size`
   - Truncate each bucket file to its `bucket_sizes[i]` value
   - If checkpoint sizes are missing/corrupt, restart process phase from scratch
4. For each system:
   - Parse (id, name, coords, star_class)
   - Normalize coordinates, compute Hilbert key
   - Append name to `names.bin`, record offset
   - Append 37-byte record to bucket file
   - Periodically update state file
5. On completion: transition state to `compile` phase

### Checkpoint Consistency

Resume correctness depends on a consistent checkpoint tuple:

- `last_system_id64`
- `names_bin_size`
- `bucket_sizes[]`

These must be written from the same checkpoint boundary (after flushing buffered writers and syncing files).

Truncating buckets to the nearest 37-byte record boundary alone is not sufficient for safe resume because it does not guarantee alignment with `last_system_id64` and `names_bin_size`.

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

> **Note:** Consider deferring all file moves (names.bin, systems.bin, systems.idx) to a final cleanup step after Phase 4 completes. This would make the cache -> data dir transition atomic and easier to reason about.

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
move cache/systems.bin  -> data/galaxy/systems.bin
move cache/systems.idx  -> data/galaxy/systems.idx
move cache/names.bin    -> data/galaxy/names.bin
move cache/names.trie   -> data/galaxy/names.trie
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

---

## Future Considerations

### Dynamic Updates

Current design is immutable (build once, query many). For dynamic updates:
- Append-only log + periodic rebuild
- Or switch to B-tree/LSM structure (significant complexity increase)

### Compression

- Records could be delta-encoded (Hilbert keys are sequential)
- Names could use dictionary compression (many shared prefixes)
- Trade-off: CPU for decompression vs I/O savings

### Memory-Mapped Access

All files are suitable for mmap:
- `systems.idx`: Small, load fully into memory
- `systems.bin`: Memory-map, let OS handle paging
- `names.trie`: Memory-map for on-demand access
- `names.bin`: Memory-map, seek to specific offsets

### 32-bit Support via Grid Partitioning

If 64-bit Hilbert keys become problematic:
- Partition galaxy into large cells (e.g., 1000 ly cubes)
- Each cell gets its own `systems.bin` with local Hilbert ordering
- Order 10 = 30 bits, fits in u32 (1024 divisions per axis = ~1 ly precision within 1000 ly cell)
- Adds two-level lookup but enables 32-bit keys within cells

### Implementation-Level Optimizations

- Parallel bucket sorting
- Streaming trie build (reduce RAM requirement)
- Incremental updates (detect changed systems)
- Compressed storage
