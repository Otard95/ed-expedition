# CSS Architecture Rules

## Table of Contents

- [Class Location Rules](#class-location-rules)
- [Naming Conventions](#naming-conventions)
- [:global() Usage Rules](#global-usage-rules)
- [Component Style Ownership](#component-style-ownership)
- [When to Extract to Global](#when-to-extract-to-global)
- [Modal Sizing Philosophy](#modal-sizing-philosophy)

---

## Class Location Rules

### Global Utilities (`style.css`)

Add to `style.css` ONLY when used in **3+ components**. Classes must earn their place.

**Allowed in global:**
- Layout: `.flex-row`, `.flex-col`, `.flex-center`, `.flex-between`, `.flex-gap-sm/md/lg`
- Typography: `.text-left`, `.text-center`, `.text-right`, `.text-uppercase-tracked`, `.numeric`
- Text colors: `.text-primary`, `.text-secondary`, `.text-dim`, `.text-danger`, `.text-warning`, `.text-success`, `.text-info`, `.text-orange`
- Animation: `.highlight`, `.blink`, `.fade-in`

### Component Styles (`<style>` block)

Everything else:
- Component-specific classes (`.route-header`, `.modal-content`)
- All classes used only in that component
- Domain-specific classes until proven widely used (3+ components)

**Rule:** Keep classes local by default. Extract to global ONLY when duplication becomes a problem.

---

## Naming Conventions

Use descriptive prefixes:

| Prefix | Affects | Examples |
|--------|---------|----------|
| `text-*` | Text styling | `.text-left`, `.text-secondary` |
| `flex-*` | Flexbox | `.flex-center`, `.flex-between` |
| `grid-*` | Grid | `.grid-center`, `.grid-span-2` |
| `bg-*` | Background | `.bg-primary`, `.bg-secondary` |
| `border-*` | Border | `.border-accent`, `.border-dashed` |

**Why:** `text-right` clearly means `text-align: right` (inline content), while `flex-end` means `justify-content: flex-end` (flex children).

---

## :global() Usage Rules

### Allowed (Rare)

**Use case:** Styling slot content you don't control.

```svelte
<!-- Table.svelte provides structure, consumer provides rows -->
<table>
  <tbody>
    {#each data as item}
      <slot {item} />
    {/each}
  </tbody>
</table>

<style>
  :global(tbody tr) {
    border-bottom: 1px solid var(--ed-border);
  }
</style>
```

### Rare Exception

**Use case:** Class forwarding for `style.css` utilities ONLY.

```svelte
<Card class="flex-center" />
```

Constraints:
- ONLY for utilities defined in `style.css`
- NOT for custom one-off styling (use component props instead)

### Forbidden

**1. Parent styling child component internals:**
```svelte
<!-- DON'T DO THIS -->
<style>
  :global(.child-internal-class) { color: red; }
</style>
```
Parent shouldn't know about child's internal structure. Use props instead.

**2. Cross-component class sharing:**
```svelte
<!-- DON'T define global class in parent for child to use -->
<style>
  :global(.shared-utility) { display: flex; }
</style>
```
Hidden dependency. Extract to `style.css` if genuinely shared.

**3. Styling your own rendered elements:**
```svelte
<!-- DON'T DO THIS - you control the element -->
<tr class="highlight">
<style>
  :global(tr.highlight) { background: orange; }
</style>
```
Use local scoped styles instead - no need for `:global()`.

### Pragmatic Exception

When props and utilities aren't logical and no better option exists:

```svelte
<IntersectionObserver class="stats-card-container" />

<style>
  :global(.stats-card-container) {
    position: sticky;
    top: 8px;
    z-index: 10;
  }
</style>
```

Use sparingly. Document why props or utilities weren't appropriate.

---

## Component Style Ownership

**Rule:** Components OWN their classes. Parents NEVER style them.

**Bad:**
```svelte
<!-- Parent styling child's classes -->
<style>
  :global(.jump-index) { color: var(--ed-text-dim); }
</style>
```

**Good:**
```svelte
<!-- Component owns its styles -->
<td class="jump-index">{index}</td>

<style>
  .jump-index { color: var(--ed-text-dim); }
</style>
```

Or if used in 3+ components, extract to `style.css`.

---

## When to Extract to Global

Move from component-local to `style.css` when:

1. **Used in 3+ components**
2. **Identical implementation** across all uses
3. **General purpose** - not specific to one feature

### Migration Process

1. Identify duplication (same class in 3+ files)
2. Verify implementations are identical
3. Extract to `style.css` with clear naming
4. Remove from components
5. Test

### Don't Extract Prematurely

- Only 1-2 components? Keep local.
- Implementations differ? Keep local.
- Feature-specific? Keep in feature component.

---

## Modal Sizing Philosophy

**Critical Rule: Modal NEVER controls its own size. Content controls size.**

Modal provides:
- Overlay styling, border, shadow, theming
- Header/footer structure
- Close button functionality

Modal deliberately omits:
- `width`, `max-width`, `height`, `max-height`

**Wrong:**
```svelte
<Modal class="my-modal">
<style>
  :global(.my-modal) { max-width: 700px; }
</style>
```

**Correct:**
```svelte
<Modal>
  <div class="my-content">
<style>
  .my-content { max-width: 700px; }
</style>
```

**Why:** Separation of concerns. Modal handles chrome, content handles layout. No fighting between modal constraints and content constraints.

This pattern applies to other container components - if designed to wrap content, it should fit the content.
