---
name: frontend
description: Svelte frontend development for this Wails app. Use when building UI components, styling, working with Wails bindings, or implementing frontend features. Covers component patterns, CSS architecture, Elite Dangerous theming, and common gotchas.
---

# Frontend Development

Svelte 3 + TypeScript + Vite, integrated with Go backend via Wails v2.

## Critical Workflow

**BEFORE writing any frontend code:**

1. **Check for existing components** - Search `frontend/src/components/` before creating anything new
2. **Check global utilities** - Review `frontend/src/style.css` for `.flex-*`, `.text-*`, etc.
3. **Keep styles component-local** - Only extract to `style.css` after used in 3+ components

**If you skip these steps, you will create duplicates and violate architecture rules.**

## Built Components

**Generic (components/):**
- Card, Badge, Button (with danger variant), ButtonLink
- Dropdown, DropdownItem, Modal, ConfirmDialog
- Table (with compact mode), TextInput, PlotterInput
- Toast (via lib/stores/toast.ts)
- Icons: Arrow, Chevron, ToggleChevron, Copy, Checkmark, CircleFilled, CircleHollow

**Feature-specific (features/):**
- expeditions/: ExpeditionCard, ExpeditionList, ExpeditionStatusBadge
- routes/: AddRouteWizard, RouteEditTable, LinkCandidatesModal
- links/: LinksSection
- fuel/: FuelAlertHandler

**Views (views/):**
- ExpeditionIndex, ExpeditionEdit, ExpeditionActive

**Utilities (lib/):**
- utils/dateFormat.ts - Date formatting helpers
- icons.ts - Centralized icon constants
- routes/edit.ts - Route/link edit wrappers
- stores/toast.ts - Toast notification store

## Key Gotchas

**Wails errors may be strings, not Error instances:**
```typescript
catch (err) {
  const msg = err instanceof Error ? err.message : String(err);
}
```

**Svelte 3 only** - No runes (`$state`, `$derived`), use `$:` for reactivity.

**Backend is source of truth** - Always reload data after mutations, don't trust local state.

**SVG attributes don't accept CSS units** - Use wrapper div for sizing:
```svelte
<!-- WRONG: width="1rem" causes error -->
<svg width={size} height={size}>

<!-- CORRECT: wrapper handles CSS units -->
<div style="width: {size}; height: {size}; display: inline-flex;">
  <svg width="100%" height="100%" viewBox="0 0 64 64">
```

See [references/patterns.md](references/patterns.md) for modal flows, auto-save patterns, and avoiding redundant computation.

## CSS Quick Reference

- **Component-local by default** - Extract to global only after 3+ uses
- **`:global()` only for slot content** - Never style child component internals from parent
- **Components own their styles** - Parents never reach into children
- **Modal fits content** - Content controls size, not modal

See [references/css-rules.md](references/css-rules.md) for complete rules, naming conventions, and examples.

## Elite Dangerous Color Palette

```css
/* Primary */
--ed-orange: #FF7800;
--ed-orange-dim: #CC6000;
--ed-orange-bright: #FFA040;

/* Backgrounds */
--ed-bg-primary: #000000;
--ed-bg-secondary: #0A0A0A;
--ed-bg-tertiary: #151515;

/* Text */
--ed-text-primary: #E0E0E0;
--ed-text-secondary: #A0A0A0;
--ed-text-dim: #707070;

/* Borders */
--ed-border: #2A2A2A;
--ed-border-accent: #FF7800;

/* Status */
--ed-status-planned: #6B7280;
--ed-status-active: #FF7800;
--ed-status-completed: #3B82F6;
--ed-status-ended: #EF4444;

/* Semantic */
--ed-success: #10B981;
--ed-warning: #F59E0B;
--ed-danger: #EF4444;
--ed-info: #3B82F6;
```

**Usage:** Orange is the star. High contrast always. Black backgrounds. Subtle borders.

## References

- **[patterns.md](references/patterns.md)** - Modal flow patterns (blocking close, hiding cancel on success), auto-save on blur, pre-computing expensive data, Wails error handling
- **[css-rules.md](references/css-rules.md)** - Class location rules, `:global()` allowed/forbidden cases, naming conventions, component style ownership, modal sizing philosophy
