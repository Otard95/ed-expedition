# FRONTEND.md

This file provides guidance for working with the ED Expedition frontend (Wails + Svelte + TypeScript).

---

## Stack

- **Framework:** Svelte 3.49.0 (not Svelte 4/5)
- **Language:** TypeScript 4.6.4
- **Bundler:** Vite 3.0.7
- **Integration:** Wails v2 (Go backend binding)
- **Package Manager:** pnpm (preferred)

## Current State

**Architecture:** View-based organization with routing. Components built following atomic design principles (generic reusables in `components/`, feature-specific in `features/`, page-level in `views/`).

**Routing:** svelte-spa-router installed for hash-based routing (required for Wails runtime injection).

**Theme:** Elite Dangerous aesthetic (deep black/dark bg, iconic orange accents, high contrast)

**Styling:** Mix of global CSS (`src/style.css`) and component-scoped styles (`<style>` blocks)

**Built Components:**
- **Generic (components/):** Card, Badge, Button, ButtonLink, Dropdown, DropdownItem, Table, Arrow
- **Feature-specific (features/expeditions/):** ExpeditionCard, ExpeditionList
- **Views (views/):** ExpeditionIndex (expedition list), ExpeditionEdit (expedition editing with route/link visualization)
- **Utilities (lib/):** Date formatting helpers (lib/utils/), icons (lib/icons.ts)

**Backend Integration:**
- GetExpeditionSummaries() - fetch expedition list
- CreateExpedition() - create new empty expedition (UUID-based, no name required)

## Directory Structure

```
frontend/
├── src/
│   ├── App.svelte              # Root component (view wrapper)
│   ├── main.ts                 # Entry point
│   ├── style.css               # Global styles (Elite Dangerous theme)
│   ├── components/             # Generic reusable UI components
│   │   ├── Card.svelte
│   │   ├── Badge.svelte
│   │   ├── Button.svelte
│   │   ├── ButtonLink.svelte  # Button-styled <a> tag for routing
│   │   ├── Dropdown.svelte
│   │   ├── DropdownItem.svelte
│   │   ├── Table.svelte       # Generic table with column alignment
│   │   └── Arrow.svelte       # SVG arrow component (direction, color props)
│   ├── features/               # Feature-specific components
│   │   └── expeditions/
│   │       ├── ExpeditionCard.svelte
│   │       └── ExpeditionList.svelte
│   ├── views/                  # Page-level components
│   │   ├── ExpeditionIndex.svelte  # Expedition list view
│   │   └── ExpeditionEdit.svelte   # Expedition editing view
│   ├── lib/
│   │   ├── utils/              # Shared utilities
│   │   │   └── dateFormat.ts   # Date formatting helpers
│   │   └── icons.ts            # Centralized icon constants (unicode + SVG)
│   └── assets/
│       ├── fonts/              # Nunito font (WOFF2)
│       ├── images/             # Wails logo
│       └── icons/              # SVG icons
│           └── Arrow.svg
├── wailsjs/                    # Auto-generated Go bindings
│   ├── go/
│   │   ├── main/App.{js,d.ts} # TypeScript-typed Go method wrappers
│   │   └── models.ts          # Generated TypeScript types from Go structs
│   └── runtime/                # Wails runtime API (clipboard, events, etc.)
├── index.html                  # HTML entry point
├── package.json                # Dependencies (versions locked)
├── vite.config.ts              # Vite configuration (minimal)
├── tsconfig.json               # TypeScript config (Svelte)
└── svelte.config.js            # Svelte preprocessor
```

## Development Commands

```bash
# Run in development mode (hot reload, watches Go backend)
wails dev

# Frontend-only dev server (no Go backend - will error on Go method calls)
cd frontend && pnpm run dev

# Type check Svelte components
cd frontend && pnpm run check

# Build production app
wails build
```

## Component Patterns

### Standard Component Structure

```svelte
<script lang="ts">
  // Imports (Wails Go methods, types, etc.)
  import {GoMethod} from '../wailsjs/go/main/App.js'

  // Props (if any)
  export let someProp: string

  // Local state
  let localState: string = "initial"

  // Functions
  function handleAction(): void {
    GoMethod(arg).then(result => {
      localState = result
    }).catch(err => {
      console.error("Error:", err)
    })
  }
</script>

<main>
  <!-- Template with component-scoped styles -->
  <div class="container">
    {localState}
  </div>
</main>

<style>
  /* Component-scoped styles (automatically scoped by Svelte) */
  .container {
    padding: 1rem;
  }
</style>
```

### Naming Conventions

- **Components:** PascalCase (e.g., `ExpeditionCard.svelte`)
- **Types:** camelCase files, PascalCase interfaces (e.g., `expedition.ts` exports `interface Expedition`)
- **Svelte stores:** camelCase (e.g., `expeditions.ts`)

## Wails Integration

### Calling Go Methods

All exported Go methods on `App` struct are auto-generated as TypeScript-typed async functions:

```typescript
// Import from generated bindings
import {MethodName} from '../wailsjs/go/main/App.js'

// Call (always returns Promise)
MethodName(arg1, arg2).then(result => {
  console.log(result)
}).catch(error => {
  console.error(error)
})
```

**Pattern:** Go method `func (a *App) MethodName(args) returnType` becomes `MethodName(args): Promise<returnType>`

### Wails Runtime API

```typescript
import {ClipboardSetText} from '../wailsjs/runtime/runtime'
import {EventsOn, EventsEmit} from '../wailsjs/runtime/runtime'

// Auto-copy to clipboard (critical for expedition next system feature)
ClipboardSetText("Beagle Point")

// Listen to Go-emitted events
EventsOn("expedition:jump", (data) => {
  console.log("New jump:", data)
})

// Emit event to Go
EventsEmit("ui:ready")
```

**Available APIs:**
- **Clipboard:** `ClipboardGetText()`, `ClipboardSetText(text)`
- **Events:** `EventsOn(event, callback)`, `EventsEmit(event, data)`, `EventsOff(event)`
- **Logging:** `LogDebug()`, `LogInfo()`, `LogError()`, `LogFatal()`
- **Window:** `WindowSetTitle()`, `WindowCenter()`, `WindowMaximise()`, etc.

## Elite Dangerous Color Palette

**Primary Colors:**
```css
--ed-orange: #FF7800;        /* Iconic ED accent/primary */
--ed-orange-dim: #CC6000;    /* Dimmed orange for hover/inactive */
--ed-orange-bright: #FFA040; /* Bright orange for highlights */

--ed-bg-primary: #000000;    /* Deep black background */
--ed-bg-secondary: #0A0A0A;  /* Slightly lighter black for cards/panels */
--ed-bg-tertiary: #151515;   /* Lighter still for nested elements */

--ed-text-primary: #E0E0E0;  /* Light gray text (high contrast) */
--ed-text-secondary: #A0A0A0;/* Dimmed text for less important info */
--ed-text-dim: #707070;      /* Very dim text for labels */

--ed-border: #2A2A2A;        /* Subtle borders */
--ed-border-accent: #FF7800; /* Orange borders for active/focused */
```

**Status Colors:**
```css
--ed-status-planned: #6B7280;   /* Gray - planned expedition */
--ed-status-active: #FF7800;    /* Orange - active expedition (use primary) */
--ed-status-completed: #3B82F6; /* Blue - completed */
--ed-status-ended: #EF4444;     /* Red - ended/failed */
```

**Semantic Colors:**
```css
--ed-success: #10B981;    /* Green - success states */
--ed-warning: #F59E0B;    /* Yellow/amber - warnings */
--ed-danger: #EF4444;     /* Red - danger/errors */
--ed-info: #3B82F6;       /* Blue - informational */
```

### Color Usage Guidelines

1. **Orange is the star** - Use `--ed-orange` for primary actions, active states, focus indicators
2. **High contrast always** - Elite Dangerous UI is readable at a glance in space
3. **Black backgrounds** - Use `--ed-bg-primary` for main background, `--ed-bg-secondary` for cards
4. **Subtle borders** - Use `--ed-border` for separation, `--ed-border-accent` for active/focus
5. **Orange glow for active** - Box shadows with orange for active expedition cards

**Example:**
```css
.card {
  background: var(--ed-bg-secondary);
  border: 1px solid var(--ed-border);
  color: var(--ed-text-primary);
}

.card.active {
  border-color: var(--ed-orange);
  box-shadow: 0 0 10px rgba(255, 120, 0, 0.3);
}

.button-primary {
  background: var(--ed-orange);
  color: #000;
}

.button-primary:hover {
  background: var(--ed-orange-bright);
}
```

## Styling Guidelines

### Use Component-Scoped Styles

```svelte
<style>
  /* Automatically scoped to this component - no conflicts */
  .card {
    padding: 1rem;
    border-radius: 8px;
  }
</style>
```

### Use CSS Variables for Theming

```svelte
<!-- Parent component -->
<Card --card-bg="rgba(40, 50, 70, 1)" />

<!-- Card.svelte -->
<style>
  .card {
    background: var(--card-bg, rgba(30, 40, 60, 1));
  }
</style>
```

### Extend Global Theme

Current global styles in `src/style.css`:
- Background: `rgba(27, 38, 54, 1)` (dark blue)
- Text: `white`
- Font: `Nunito` (loaded locally)
- Layout: Center-aligned, full viewport height

**Add global styles sparingly.** Prefer component-scoped styles.

## TypeScript Types

### Define Types Matching Go Structs

```typescript
// src/lib/types/expedition.ts
export type ExpeditionStatus = "planned" | "active" | "completed" | "ended"

export interface Expedition {
  id: string
  name: string
  status: ExpeditionStatus
  created_at: string
  last_updated: string
  // Match Go struct fields exactly
}
```

### Use Explicit Type Annotations

```typescript
let expeditions: Expedition[] = []
let loading: boolean = false
let error: string | null = null

function fetchExpeditions(): void {
  loading = true
  ListExpeditions().then((result: Expedition[]) => {
    expeditions = result
    loading = false
  })
}
```

## Error Handling Pattern

```typescript
import {GetExpedition} from '../wailsjs/go/main/App.js'

let data: Expedition | null = null
let loading: boolean = true
let error: string | null = null

function loadExpedition(id: string): void {
  loading = true
  error = null

  GetExpedition(id)
    .then((result: Expedition) => {
      data = result
      loading = false
    })
    .catch((err: Error) => {
      error = err.message
      loading = false
    })
}
```

## Svelte 3 Specifics

**Reactivity syntax:** Use `$:` for reactive statements

```typescript
let count: number = 0
$: doubled = count * 2  // Automatically updates when count changes
```

**Don't use Svelte 4/5 syntax:**
- No runes (`$state`, `$derived`, `$effect`)
- No `<svelte:component>` with `this={}`
- Use Svelte 3 patterns only

## Asset Imports

```typescript
import logo from './assets/images/logo.png'
import font from './assets/fonts/font.woff2'

// Use in template
<img src={logo} alt="Logo" />
```

Vite handles asset bundling and optimization automatically.

## Common Patterns for Expedition UI

### List Expeditions

```typescript
import {ListExpeditions} from '../wailsjs/go/main/App.js'

let expeditions: Expedition[] = []

ListExpeditions().then(result => {
  expeditions = result
})
```

### Display Status Badge (Elite Dangerous Style)

```svelte
<script lang="ts">
  export let status: ExpeditionStatus

  const colors = {
    planned: "#6B7280",   // Gray
    active: "#FF7800",    // ED Orange
    completed: "#3B82F6", // Blue
    ended: "#EF4444"      // Red
  }
</script>

<span class="badge" class:active={status === 'active'} style="background-color: {colors[status]}">
  {status.toUpperCase()}
</span>

<style>
  .badge {
    padding: 0.25rem 0.5rem;
    border-radius: 2px;
    font-size: 0.75rem;
    font-weight: 600;
    color: #000;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .badge.active {
    box-shadow: 0 0 8px rgba(255, 120, 0, 0.5);
  }
</style>
```

### Format Dates

```typescript
function formatDate(isoString: string): string {
  const date = new Date(isoString)
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  }).format(date)
}

function formatRelativeTime(isoString: string): string {
  const date = new Date(isoString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffDays === 0) return "Today"
  if (diffDays === 1) return "Yesterday"
  if (diffDays < 7) return `${diffDays} days ago`
  return formatDate(isoString)
}
```

## Critical Constraints from SPEC

1. **Single active expedition** - Only one expedition can be `active` at a time
2. **Auto-copy to clipboard** - Use `ClipboardSetText()` for next system name
3. **Routes are immutable** - Never allow editing route data
4. **Four states** - planned → active → completed/ended (no reverse transitions)
5. **Circular routes** - Need to handle `baked_loop_back_index` for visual indication

## Completed Work

✅ **Generic Components (components/)**
- Card.svelte - Styled container with variant support (default/active)
- Badge.svelte - Status indicators with ED color palette
- Button.svelte - Primary/secondary button variants
- ButtonLink.svelte - Button-styled `<a>` tag for routing
- Dropdown.svelte - Three-dot menu with click-outside detection
- DropdownItem.svelte - Dropdown menu items
- Table.svelte - Generic table with column alignment, data iteration, slot-based cell content
- Arrow.svelte - SVG arrow component with direction/color/size props

✅ **Feature Components (features/expeditions/)**
- ExpeditionCard.svelte - Displays expedition summary with actions
- ExpeditionList.svelte - Vertical list with empty state

✅ **Views (views/)**
- ExpeditionIndex.svelte - Expedition list view (calls GetExpeditionSummaries)
- ExpeditionEdit.svelte - Expedition editing view with route/link visualization
  - Displays routes as vertical tables (Card + Table component)
  - Link badges (incoming=blue, outgoing=orange) with click-to-scroll navigation
  - Target row highlight and blink animations on navigation
  - Mock data for testing UI

✅ **Utilities (lib/)**
- dateFormat.ts - Locale-aware date formatting (formatDate, formatRelativeTime)
- icons.ts - Centralized icon constants (ARROW_RIGHT, ARROW_LEFT, SCOOPABLE, NOT_SCOOPABLE)

✅ **Backend Integration**
- GetExpeditionSummaries() - fetch expedition list
- CreateExpedition() - create new empty expedition (UUID-based)

✅ **Routing Setup**
- svelte-spa-router installed (hash-based routing for Wails)
- View-based architecture ready for routing implementation

## Next Development Steps

- Implement actual routing with svelte-spa-router (connect views to routes)
- Route import UI (paste Spansh JSON, save to route library)
- Create expedition flow (select routes, define links, set start point)
- Replace mock data in ExpeditionEdit with real backend calls
- FSDJump processing (baked route progression, clipboard auto-copy)
- State management for active expedition
- Additional expedition management actions (start, end, delete)
