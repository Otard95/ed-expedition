# Systems Database Specification

This document specifies the binary file formats for storing and querying ~170 million star systems with spatial indexing and name lookup.

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

> **Note:** Multiple systems may share the same Hilbert index due to 0.1 ly quantization (minimum system distance is 0.03125 ly). This is fine - Hilbert key is a sort key, not a unique identifier.

> **Tuning:** Order 21 (0.05 ly precision) uses 63 bits and reduces collisions. Order 19 (0.2 ly precision) uses 57 bits if smaller keys are needed.

---

## systems.bin

Fixed-size records sorted by Hilbert key. Enables spatial queries via binary search + range scan.

### Record Format (33 bytes)

| Field | Type | Bytes | Description |
|-------|------|-------|-------------|
| hilbert_key | u64 | 8 | Hilbert curve index (order 20, 60 bits used) |
| x | u32 | 4 | Normalized X coordinate, scaled by 10 |
| y | u32 | 4 | Normalized Y coordinate, scaled by 10 |
| z | u32 | 4 | Normalized Z coordinate, scaled by 10 |
| id | u64 | 8 | System id64 (Elite Dangerous unique identifier) |
| star_class | u8 | 1 | Star classification enum (see below) |
| name_offset | u32 | 4 | Byte offset into names.bin |

**Size:** ~170,000,000 systems × 33 bytes = **~5.61 GB**

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

> **Tuning:** Scan range depends on query radius and local density. For typical 50 ly jump range queries, scanning ±10,000 systems from the estimated position should be more than sufficient.

---

## systems.idx

Sparse index mapping Hilbert keys to file offsets. Loaded into memory at startup.

### Format

Array of entries, one per bucket (every 50,000 systems):

| Field | Type | Bytes | Description |
|-------|------|-------|-------------|
| hilbert_key | u64 | 8 | Hilbert key of first system in bucket |
| file_offset | u64 | 8 | Byte offset into systems.bin |

**Entry count:** ceil(~170,000,000 / 50,000) = **~3,400 entries**

**Size:** ~3,400 × 16 bytes = **~54.4 KB**

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

---

## names.trie

Prefix trie for name lookup and autocomplete. Optimized for minimal seeks during lookup.

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
| str_len | u16 | 2 | Length of segment string |
| segment | [u8] | str_len | The name segment (e.g., "Lyruewry", "AA-A", "h0") |
| flags | u8 | 1 | Bit 0: is_terminal (rest reserved) |
| ref | u64 | 8 | If terminal: byte offset into systems.bin. If non-terminal: byte offset to first child's sibling group |

**Node size:** 11 + str_len bytes

### Structure

- File starts with root sibling group at offset 0
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
3. Match found, non-terminal → seek to children_offset (100)
4. Read sibling_count, sequential scan until "AA-A" matches
5. Match found, non-terminal → seek to children_offset (300)
6. Read sibling_count, sequential scan until "h1" matches
7. Match found, terminal → return system_offset

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

### Format

```
<name1>\0<name2>\0<name3>\0...
```

No structure, no length prefixes. Each name is null-terminated. `name_offset` points to the first character of the name.

**Size estimate:** ~170,000,000 systems × ~18 chars average = **~3.1 GB**

> **Note:** Ordering doesn't matter. Names can be written in any order during build - each system record stores its specific offset.

> **Note:** This duplicates name data that's also in `names.trie`. The ~3 GB cost is accepted for simplicity - reconstructing names from the trie would require parent pointers or path tracking, adding complexity, and crucially too much IO/seek in `names.trie`.

---

## Build Process

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

## Query Patterns

### Find system by exact name

```
1. Traverse names.trie to terminal node
2. Read system_offset from terminal node
3. Seek to system_offset in systems.bin
4. Read 33-byte record
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
