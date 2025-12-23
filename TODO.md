# TODO

This file tracks known issues, technical debt, and planned features for the ED Expedition app.

---

## 🔴 CRITICAL - Must Fix Before Shipping

### Active Expedition View - Missing First System

**Issue:** Active expedition view didn't show the first system when an expedition was newly started.

**Probable cause found (Dec 2025):**
- When `CurrentBakedIndex` was initialized to 0, the slice logic `bakedRoute.jumps.slice(expedition.current_baked_index + 1)` would start at index 1, skipping the first system
- With new -1 initialization, slice starts at index 0 when expedition not started yet, showing all systems including first
- **Fix:** CurrentBakedIndex now starts at -1 when location unknown, 0 only if at start system with synthetic jump
- **Status:** Probably fixed - needs verification with real gameplay test

**Files involved:**
- `services/expedition_lifetime.go` - CurrentBakedIndex initialization
- `frontend/src/lib/expedition/active.ts` - Slice logic in computeActiveStats

---

## 🟡 High Priority

### Investigate First Jump App State Save Failure

**Issue:** App state save fails on first FSD jump of session/expedition, causing crash.

**Symptoms:**
- Crash occurs after first FSD jump
- Error: Failed to save app state (likely nil pointer exception)
- Stack trace stopped at `database/json.go:28` - `json.MarshalIndent(data, "", "  ")`
- Only happens on first jump (subsequent jumps work fine)
- Unsure if "first of expedition" vs "first of session" matters

**What we know:**
- Stack trace indicates `json.MarshalIndent` fails when marshaling AppState
- Suggests nil pointer or unmarshallable field in AppState structure
- Crash was in `AppStateService` when saving after FSD jump event
- Commit `402ac90` replaced panic with error logging for better diagnostics
- Next occurrence should provide detailed error logs

**Probable cause found (Dec 2025):**
- `StartExpedition` was not setting `e.bakedRoute = route` after creating the baked route
- After starting an expedition, `e.bakedRoute` remained nil (or stale from previous expedition)
- First jump after start would have `handleJump` panicking from nil pointer dereference
- The `json.MarshalIndent` in the crash output was likely a red herring - Go dumps all goroutine stacks on panic, and AppStateService happened to be mid-marshal when `handleJump` panicked
- **Fix:** Added `e.bakedRoute = route` after successful transaction apply in `StartExpedition`
- **Status:** Probably fixed - needs verification with real gameplay test

**Files involved:**
- `services/expedition_lifetime.go` - StartExpedition was missing `e.bakedRoute` assignment
- `database/json.go:28` - JSON marshal failure point
- `services/app_state.go:70` - FSD jump event handler with save
- `models/app_state.go` - AppState struct definition and SaveAppState

**Temporary mitigation:** Error logging instead of panic (commit `402ac90`)

### ~~Implement Transaction/Rollback System~~ ✅ DONE

**Status:** Completed (commit e0d1781 - Jan 2025)

**Implementation:** `database.Transaction` provides atomic multi-file writes with automatic rollback on failure. All multi-step operations (create expedition, add route, start expedition) now use transactions.

**Remaining enhancements:** See Technical Debt section for logging and startup recovery improvements.

---

## 🟢 Medium Priority

### Improve Completed Expedition View Design

**Current state:** Basic stats grid + simple jump list.

**Goal:** Mix of edit and active view patterns:
- Grouped stats (Time, Jumps, Distance) like completion modal
- Table component for jump history (consistent with active view)
- Show started/ended dates, duration, accuracy %

**Files:**
- `frontend/src/views/ExpeditionView.svelte`

### Visual Polish

- Better loading states during backend operations
- Toast notifications instead of alerts
- **Fix AddRouteWizard Step 3 loading spinner** - Spinner not centered during route plotting

### CustomSelect Dropdown UX

**Issue:** CustomSelect dropdown doesn't auto-close when hovering away (unlike Dropdown component in route edit tables).

**Current behavior:**
- Click outside closes dropdown (implemented)
- Escape key closes dropdown (implemented)
- No hover-timeout to auto-close

**Enhancement:** Add hover-timeout like Dropdown component - if not hovered for X seconds, auto-close.

**Files:**
- `frontend/src/components/CustomSelect.svelte`
- Reference: `frontend/src/components/Dropdown.svelte` for hover behavior

### Track In-Game Fuel Level and Warn on Low Fuel

**Goal:** Track current fuel level from journal events and warn user if fuel is lower than route expects for upcoming jumps.

**Implementation needed:**

1. **Backend - Fuel tracking service:**
   - Monitor journal events: `Refuel`, `FuelScoop`, `RefuelAll`, `RefuelPartial`
   - Track fuel consumption from `FSDJump` events (fuel_used field)
   - Maintain current fuel state in app state
   - Expose current fuel level to frontend

2. **Backend - Fuel warning logic:**
   - Compare current fuel against next N jumps in baked route (e.g., next 3-5 jumps)
   - Calculate fuel required for upcoming segment (sum of fuel_used)
   - Account for scoopable stars in the segment (can refuel at those)
   - Return warning status: safe, low (< 2 jumps worth), critical (< 1 jump)

3. **Frontend - Fuel display:**
   - Show current fuel level in active expedition view
   - Display fuel warning badge/indicator when low
   - Optionally: Show "fuel range" (how many jumps possible with current fuel)

4. **Frontend - Warning UI:**
   - Visual warning indicator when fuel is insufficient
   - Toast notification when fuel drops below safe threshold
   - Highlight next scoopable star in route table

**Challenges:**
- Initial fuel level unknown until first Refuel/FuelScoop event (start with max fuel capacity assumption?)
- Need ship's max fuel capacity from Loadout event
- Must handle fuel tanks in Loadout (standard + optional tanks)
- Edge case: Player refuels at stations (not just scooping) - both need tracking

**Files affected:**
- `journal/events.go` - Add Refuel event types
- `models/app_state.go` - Add current fuel level field
- `services/app_state.go` - Track fuel from journal events
- `services/expedition.go` - Add fuel warning calculation method
- `app.go` - Expose fuel status to frontend
- `frontend/src/views/ExpeditionActive.svelte` - Display fuel status and warnings

**Benefits:**
- Prevents running out of fuel during expeditions
- Proactive warning system for fuel management
- Highlights next refuel opportunity (scoopable star)

---

## 🔵 Low Priority / Future

### Custom Scroll Animation for Active View

**Current behavior:**
- Active view uses `scrollIntoView({ behavior: "smooth" })` to scroll to current jump
- Browser's default smooth scroll - no control over speed or easing
- Animation is browser-dependent and not customizable

**Enhancement:**
- Implement custom scroll animation using `requestAnimationFrame`
- Add configurable duration (e.g., 500ms)
- Add custom easing functions (easeInOutQuad, easeOutCubic, etc.)
- More polished, consistent animation across browsers

**Implementation approach:**
- Create utility function with manual animation loop
- Support easing parameter (string or function)
- Replace `scrollIntoView` call in ExpeditionActive.svelte

**Priority:** Nice to have, not important

**Files affected:**
- `frontend/src/views/ExpeditionActive.svelte:110-117` - Current scroll implementation
- Possibly new utility file for reusable scroll animation

### Additional Completion Stats

Consider adding these stats to the expedition completion modal:

- **Total Fuel Used** - Sum of fuel_used from all jumps (resource tracking)
- **Shortest Jump** - Minimum jump distance (might highlight emergency refuels or detours)
- **Average Fuel per Jump** - Shows efficiency over the expedition
- **Number of Unexpected Jumps** - On-route jumps that weren't the expected next system (rerouting count)
- **Average Time Between Jumps** - Shows pacing/speed (useful for expeditions lasting hours/days)
- **Unique Systems Visited** - Count of unique system_ids vs total jumps (shows if systems were revisited)

**Current stats shown:**
- Duration, Total Jumps, Route Accuracy %, Total Distance, Average Jump, Longest Jump

**Decision needed:** Which (if any) would add value without cluttering the completion screen?

### Detour Icon in Active View Status Column

**Current behavior:**
- On-route jumps show the Route icon
- Off-route/detour jumps show nothing (empty status indicator)

**Consideration:**
- May want a visual indicator for detour jumps to make them more obvious
- Unclear if this is necessary or what icon would work best

**Possible icon concepts:**
- Curved/bent arrow (↪) - Universal detour symbol
- Diagonal slash (/) - Simple off-route indicator
- Warning triangle - ED-style deviation warning
- Hollow/outlined route icon - "Jump but not on route"
- Zigzag line - Erratic path

**Decision needed:**
- Is a detour icon actually useful, or is the empty space sufficient?
- If yes, what icon style fits the ED theme and is distinct from chevron/target/route?

**Files affected:**
- New icon component (e.g., `frontend/src/components/Detour.svelte`)
- `frontend/src/features/routes/RouteActiveTable.svelte` - Add detour icon rendering

### Multiple Active Expeditions

Current design: only one active expedition at a time. Future enhancement could allow multiple active expeditions.

**Requires:**
- Index.json schema change
- UI to switch between active expeditions
- Journal watcher would need to route FSDJump events to correct expedition

### Auto-Create Link After "Link to New Route"

**Current behavior:**
- "Link to new route" opens AddRouteWizard with pre-filled start system
- User plots route, route gets added
- User must manually create the link afterward

**Enhancement:**
- After route is added via "Link to new route", automatically:
  1. Find the jump in the new route that matches the start system
  2. Create the link from the original jump to the new route's matching jump
  3. Determine link direction based on original context (from/to)

**Challenges:**
- Need to remember which jump triggered the "Link to new route" flow
- Need to determine which jump in the new route to link to (first occurrence of system)
- Need to handle case where system doesn't exist in new route (unlikely but possible)

**Files affected:**
- `frontend/src/views/ExpeditionEdit.svelte` - Track original jump context
- `frontend/src/features/routes/AddRouteWizard.svelte` - Pass context through wizard
- Backend might need a "create link after route add" helper

### Migrate to SQLite

**Rationale:** JSON file storage with manual transaction system has inherent limitations. SQLite provides real ACID transactions, eliminating the partial-apply problem entirely.

**When:** Once the app is more stable and feature-complete. Not worth the migration effort while data model is still evolving.

**Benefits:**
- True atomic transactions (no partial-apply failures)
- No orphaned files or inconsistent state possible
- Better query capabilities for future features
- Handles concurrent access properly

**Scope:**
- Replace `database/json.go` with SQLite wrapper
- Migrate all models to use SQLite storage
- Data migration tool for existing JSON files
- Remove transaction system (SQLite handles this)

**Files affected:**
- `database/` - Complete rewrite
- All `models/*.go` - Update Load/Save functions
- Possibly add migration tool in `cmd/`

### Multiple Links Per Jump Support

**Current constraint:** Maximum one link per jump (simplified for v1).

**Limitation for complex cycles:**
Some cycle patterns require multiple links on the same system (e.g., system appears 3+ times across routes with different connections). Current constraint means these patterns require workarounds or are impossible.

**Example scenario that would need multiple links per jump:**
```
Route 1: a -> b
Route 2: b -> c -> d
Route 3: d -> e -> b

Desired links:
- Link 1: Route 1 Jump 2 -> Route 2 Jump 1 (on system b)
- Link 2: Route 2 Jump 3 -> Route 3 Jump 1 (on system d)
- Link 3: Route 3 Jump 3 -> Route 2 Jump 1 (on system b) ⚠️ conflicts with Link 1
```

System b would need 2 incoming links, which current design doesn't support.

**To support multiple links per jump:**
1. Backend: Allow multiple links per jump (validate: max 1 outgoing, unlimited incoming)
2. Frontend: Change `EditViewRoute.jump.link` to `links: EditViewLink[]`
3. UI: Render multiple link badges per jump
4. UI: Update dropdown conditional to check for outgoing in array

**Why deferred:** Cycles are rare use case, most players won't hit this limitation.

**Files affected:**
- Backend validation logic (currently unknown if enforced)
- `frontend/src/lib/routes/edit.ts` - EditViewRoute decoration
- `frontend/src/features/routes/RouteEditTable.svelte:265` - Dropdown conditional (currently `{#if !item.link}`)

---

## Technical Debt

### Transaction System - Add Logging

**Enhancement:** Pass a logger into Transaction for debugging.

**Use cases:**
- Log when transaction is created (with id/name)
- Log each WriteJSON staging
- Log Apply/Rewind with success/failure details

**Files:** `database/json.go`

### Transaction System - Startup Recovery

**Remaining issue:** If `Apply()` fails partway through (rare but possible), the app is left in an inconsistent state.

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

---

- Error handling needs improvement (currently using alerts, should use proper error UI)
- No tests (unit or integration)
- No TypeScript strict mode
- Frontend build warnings need cleanup
