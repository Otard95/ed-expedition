# TODO

This file tracks known issues, technical debt, and planned features for the ED Expedition app.

---

## 🔴 CRITICAL - Must Fix Before Shipping

(No critical issues currently)

---

## 🟡 High Priority

### Implement Transaction/Rollback System

**Issue:** Multi-step operations (create expedition, add route, start expedition) can fail partway through, leaving orphaned files or inconsistent state between index and expedition files.

**Impact:** Data corruption risk, manual cleanup required if failures occur during save operations.

**Solution needed:** Implement atomicity for multi-step operations - either all changes succeed or all are rolled back.

**See:** Detailed breakdown in Technical Debt section below.

---

## 🟢 Medium Priority

### Preserve Route Collapse State

**Problem:** Route collapse state resets when creating links (annoying UX).

**Current behavior:**
- User expands/collapses routes manually
- Create a link → expedition data reloads
- All routes reset to default collapsed state
- User has to re-expand routes they were working with

**Solution:**
- Store collapse state in local component state or session storage
- Restore collapse state after expedition reload
- Key by route ID to handle route additions/deletions

**Files affected:**
- `frontend/src/views/ExpeditionEdit.svelte` - Parent managing route list
- `frontend/src/features/routes/RouteEditTable.svelte` - Individual route collapse state

### Visual Polish

- Better loading states during backend operations
- Toast notifications instead of alerts

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

### Transaction/Rollback Handling

**Problem:** No atomicity in multi-step operations - failures can leave orphaned data or inconsistent state.

**Current issues:**

1. **Orphan expeditions** (`services/expedition_lifetime.go:81`)
   - `CreateExpedition()` saves expedition, then saves index
   - If index save fails, expedition file is orphaned

2. **Orphan routes** (`services/expedition_edit.go:18`)
   - `AddRouteToExpedition()` saves route, then updates expedition
   - If expedition update fails, route file is orphaned

3. **Orphan baked routes** (`services/expedition_lifetime.go:174`)
   - `StartExpedition()` saves baked route, then updates expedition
   - If expedition save fails, baked route file is orphaned

4. **Inconsistent index** (`services/expedition_lifetime.go:200`)
   - `StartExpedition()` saves expedition with active status, then updates index
   - If index save fails, expedition is active on disk but index doesn't reflect it

**Solution approaches:**
- Implement transaction log/journal for multi-step operations
- Add cleanup/recovery logic on startup to detect and fix orphans
- Use a proper database instead of JSON files (future consideration)

**Why deferred:** Rare edge cases, manual recovery possible for now. Proper fix requires significant refactoring.

---

- Error handling needs improvement (currently using alerts, should use proper error UI)
- No tests (unit or integration)
- No TypeScript strict mode
- Frontend build warnings need cleanup
