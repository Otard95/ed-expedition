# Frontend Patterns

## Table of Contents

- [Component Philosophy](#component-philosophy)
- [Modal Flow for Backend Mutations](#modal-flow-for-backend-mutations)
- [Wails Error Handling](#wails-error-handling)
- [Auto-Save on Blur](#auto-save-on-blur)
- [Avoid Redundant Computation](#avoid-redundant-computation)
- [Standard Component Structure](#standard-component-structure)

---

## Component Philosophy

**Simple, Generic, Composable**

Reusable components provide styling primitives, not layout decisions.

**Rules:**
1. **Generic components handle styling only** - Padding, border, shadow, colors, variants. Accept `class` prop for overrides.
2. **Content consumers control layout** - Parent decides flexbox, grid, spacing. Components use default slot without layout opinions.
3. **Atomic design hierarchy:**
   - **Atoms**: Card, Button, Badge (pure styling)
   - **Molecules**: ExpeditionCard (simple compositions)
   - **Organisms**: ExpeditionList (complex features)
   - **Pages**: Route-level views

**Example - Card should be a styled box, not impose layout:**
```svelte
<!-- GOOD: Consumer controls layout -->
<div class="card {variant} {className}">
  <slot />
</div>

<!-- BAD: Hardcoded internal layout -->
<div class="card" style="display: flex; flex-direction: column; gap: 1rem;">
  <div class="header"><slot name="header" /></div>
  <div class="body"><slot /></div>
</div>
```

---

## Modal Flow for Backend Mutations

When building multi-step modals that perform backend mutations:

### 1. Block close during async operations

```typescript
$: canClose = currentStep !== "plotting" && currentStep !== "success";
```

### 2. Hide Cancel/Back on success step

Force users to click "Finish" for proper cleanup:

```typescript
$: showCancel = currentStep !== "plotting" && currentStep !== "success";
$: showBack = currentStep !== "select-plotter" && currentStep !== "plotting" && currentStep !== "success";
```

### 3. Reload from backend after mutations

Go is source of truth, file I/O is cheap:

```typescript
async function handleComplete() {
  await MutateBackend(id, data);
  expedition = await LoadExpedition(id);
  expeditionName = expedition.name || "";
}
```

### 4. Pass callbacks for parent updates

```svelte
<Modal>
  <Wizard onComplete={handleReload} />
</Modal>
```

---

## Wails Error Handling

Wails errors may be strings, not Error instances:

```typescript
try {
  await WailsMethod();
} catch (err) {
  const errorMsg = err instanceof Error
    ? err.message
    : typeof err === 'string'
      ? err
      : String(err);
  alert(`Failed: ${errorMsg}`);
}
```

---

## Auto-Save on Blur

Pattern for inputs that save automatically when focus leaves:

```svelte
<script>
  let value = "";
  let saving = false;

  async function handleBlur() {
    if (saving) return;
    const trimmed = value.trim();
    if (trimmed === originalValue) return;

    saving = true;
    try {
      await SaveValue(id, trimmed);
      data = await LoadData(id);
      value = data.value || "";
    } catch (err) {
      console.error("Save failed:", err);
      value = originalValue; // Revert on failure
    } finally {
      saving = false;
    }
  }
</script>

<input bind:value on:blur={handleBlur} disabled={saving} />
```

Key points:
- Guard against concurrent saves with `saving` flag
- Skip save if value unchanged
- Reload from backend after save (source of truth)
- Revert to original on failure

---

## Avoid Redundant Computation

**Problem:** Computing the same data in parent and child components.

**Antipattern:**
```typescript
// Parent computes candidates
$: possibleCandidates = computeExpensiveThing(allData);

// Modal re-computes the same data
function findCandidates(allData, filter) {
  // Iterates through allData again...
}
```

**Solution:** Pre-compute in parent, pass subset to child:
```typescript
// Parent computes once
$: possibleCandidates = computeExpensiveThing(allData);

// Pass pre-computed subset to child
<Modal candidates={possibleCandidates[filterId]} />
```

**When to care:** Large datasets (100+ items), nested loops, or data used reactively. Don't optimize prematurely for small datasets.

---

## Standard Component Structure

```svelte
<script lang="ts">
  // Imports
  import { GoMethod } from '../wailsjs/go/main/App.js';

  // Props
  export let someProp: string;

  // Local state
  let localState: string = "initial";

  // Functions
  function handleAction(): void {
    GoMethod(arg).then(result => {
      localState = result;
    }).catch(err => {
      console.error("Error:", err);
    });
  }
</script>

<main>
  <div class="container">
    {localState}
  </div>
</main>

<style>
  .container {
    padding: 1rem;
  }
</style>
```

### Naming Conventions

- **Components:** PascalCase (`ExpeditionCard.svelte`)
- **Type files:** camelCase (`expedition.ts`)
- **Interfaces:** PascalCase (`interface Expedition`)
- **Stores:** camelCase (`expeditions.ts`)
