# Galaxy Database - Design Decisions

This document captures the reasoning behind key design choices. See `systems.md` for the spec and `systems-impl.md` for implementation details.

---

## Spatial Indexing: Hilbert Curve

### Why Hilbert over other approaches?

**Considered:**
- **Naive 3D grid with HashMap** - Simple, but poor locality for range queries
- **Z-order curve (Morton)** - Simpler math, but worse locality preservation than Hilbert
- **Octree/KD-tree** - Good for dynamic data, but complex and pointer-heavy
- **R-tree** - Optimized for rectangles, overkill for point data

**Chosen: Hilbert curve** because:
- Best locality preservation of any space-filling curve
- Maps 3D → 1D while keeping nearby points mostly nearby
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

**Constraint:** 64-bit keys, 3D space → max 21 bits per axis (63 bits total)

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

**Design:** Sample every 50,000 systems → 3,400 index entries → ~55 KB

**Lookup:**
1. Find bucket containing target key (scan 3,400 entries)
2. Lerp within bucket to estimate position
3. Binary search from estimate

**Why lerp works within buckets:** 50,000 systems in a bucket span a small, localized region of space. Density is relatively uniform at this scale, so linear interpolation gets close.

**Result:** Effectively O(1) bucket lookup + ~5-10 binary search iterations instead of ~27.

### Why not HashMap for bucket lookup?

Considered: `map[uint64]uint64` keyed by bucket's first Hilbert key.

**Problem:** You need to find which bucket *contains* a key, not exact key match. Would need to find largest key ≤ target. HashMap doesn't support this.

**Solution:** Sorted slice, linear scan or binary search. 3,400 entries is tiny - linear scan is nanoseconds.

---

## Build Pipeline: Three Separate Phases

### Why not stream directly from HTTP to final files?

**Original idea:** Download gzipped JSON → decompress → parse → transform → write final files, all streaming.

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
- 1000 buckets × 170k systems/bucket
- 170k × 37 bytes = ~6.3 MB per bucket - easily fits in memory
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
- systems.bin: 170M × 37 bytes = 6.3 GB
- names.bin: 170M × ~18 chars = 3.1 GB
- names.trie: ~500 MB - 1 GB
- systems.idx: ~55 KB

**Could reduce via:**
- Compression - adds CPU overhead, complicates random access
- Delta encoding - modest savings, complexity cost
- Smaller types - already using u32 for coords, u8 for star_class

**Decision:** Accept 10 GB. Storage is cheap, simplicity is valuable. Compression can be added later if needed.

### Why require ~8 GB RAM for trie build?

Trie construction needs all names in memory (sorted). 170M names × ~50 bytes overhead = ~8 GB.

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
