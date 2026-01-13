# TODO

Internal developer notes and technical debt. For bugs and features, see [GitHub Issues](https://github.com/Otard95/ed-expedition/issues).

---

## Technical Debt

### Transaction System - Add Logging

Pass a logger into Transaction for debugging.

**Use cases:**
- Log when transaction is created (with id/name)
- Log each WriteJSON staging
- Log Apply/Rewind with success/failure details

**Files:** `database/json.go`

### Transaction System - Startup Recovery

If `Apply()` fails partway through (rare but possible), the app is left in an inconsistent state.

**Needed:**
1. On startup, detect inconsistent state:
   - Expedition status doesn't match index
   - Orphaned .tmp files in data directory
   - Baked route exists but expedition not marked active
2. Prompt user or auto-repair

**Known orphan scenarios (for detection):**
- Orphan expeditions: expedition file exists but not in index
- Orphan routes: route file exists but not in any expedition
- Orphan baked routes: baked route file exists but expedition not active
- Inconsistent index: expedition marked active but index.ActiveExpeditionID is nil (or vice versa)

**Files:**
- `database/json.go` - Transaction implementation
- New: startup recovery logic (location TBD)

### Optimize Single-Subscriber FanoutChannels

**When:** After app is stable and relatively feature complete.

`channels.FanoutChannel` is designed for multiple subscribers (pub/sub pattern), but some channels may only ever have one subscriber. Using FanoutChannel for single-subscriber cases adds unnecessary overhead.

**Task:**
1. Audit all `FanoutChannel` usages in `journal/watcher.go`
2. Identify channels that only have one subscriber
3. Replace single-subscriber FanoutChannels with plain `chan` for better performance

**Files:**
- `journal/watcher.go` - FanoutChannel declarations
- `lib/channels/fanout.go` - FanoutChannel implementation
- Services that subscribe to channels

---

## Small Fixes

### Visual Polish

- Better loading states during backend operations
- Fix AddRouteWizard Step 3 loading spinner - not centered during route plotting

### CustomSelect Dropdown UX

CustomSelect dropdown doesn't auto-close when hovering away (unlike Dropdown component in route edit tables).

**Enhancement:** Add hover-timeout like Dropdown component - if not hovered for X seconds, auto-close.

**Files:**
- `frontend/src/components/CustomSelect.svelte`
- Reference: `frontend/src/components/Dropdown.svelte` for hover behavior

---

## Ongoing / Vague

- Error handling needs improvement (currently using alerts, should use proper error UI)
- No tests (unit or integration)
- No TypeScript strict mode
- Frontend build warnings need cleanup

---

## Consider / Revisit

### Additional Completion Stats

Consider adding to the expedition completion modal:
- Total Fuel Used, Shortest Jump, Average Fuel per Jump
- Number of Unexpected Jumps, Average Time Between Jumps
- Unique Systems Visited

**Decision needed:** Which (if any) would add value without cluttering the completion screen?

### Detour Icon in Active View

Off-route/detour jumps currently show nothing in the status column. May want a visual indicator.

**Decision needed:** Is a detour icon useful, or is empty space sufficient? If yes, what icon fits the ED theme?

### Migrate to SQLite

JSON file storage with manual transaction system has inherent limitations. SQLite provides real ACID transactions.

**When:** Once the app is more stable and data model stops evolving. Not worth the migration effort now.

---

## Internal Documentation

### Consolidate Human-Facing Frontend Documentation

Frontend docs were migrated to `.opencode/skill/frontend/` for agent consumption.

**Task:** Rewrite or consolidate `frontend/FRONTEND.md` and `frontend/CSS_RULES.md` into human-friendly documentation. Options:
1. Merge relevant parts into `frontend/README.md`
2. Rewrite as condensed human-readable docs
3. Delete if skill serves both purposes adequately

**Files:**
- `frontend/README.md` - Existing boilerplate (keep for humans)
- `frontend/FRONTEND.md` - Detailed frontend docs (needs consolidation)
- `frontend/CSS_RULES.md` - CSS architecture rules (needs consolidation)
