# CSS Architecture Rules

This document defines CSS organization, naming conventions, and usage patterns for the frontend codebase.

---

## 1. Class Location Rules

### Global Utilities (`style.css`)

Add classes to `style.css` ONLY when they are used in **3+ components**. Classes must earn their place.

**Allowed in global:**
- **Layout utilities:** `.flex-row`, `.flex-col`, `.flex-center`, `.flex-between`, `.flex-gap-sm/md/lg`
- **Typography utilities:** `.text-left`, `.text-center`, `.text-right`, `.text-uppercase-tracked`, `.numeric`
- **Semantic text colors:** `.text-primary`, `.text-secondary`, `.text-dim`, `.text-danger`, `.text-warning`, `.text-success`, `.text-info`, `.text-orange`
- **Animation utilities:** `.highlight`, `.blink`, `.fade-in`
- **Domain utilities:** Elite Dangerous-specific classes, but ONLY after they appear in 3+ components

**Examples of domain classes that would qualify:**
- `.jump-index` - If used in RouteEditTable, LinkCandidatesModal, and RoutePreview
- `.scoopable` - If used in multiple route-related components

### Component Styles (`ComponentName.svelte <style>`)

**Everything else goes here:**
- Component-specific classes (`.route-header`, `.modal-content`, `.card-body`)
- All classes used ONLY in that component
- Domain-specific classes until they prove widely used (3+ components)

**Rule:** Keep classes local by default. Extract to global ONLY when duplication becomes a problem.

---

## 2. Naming Conventions

Use descriptive prefixes to indicate what the class affects:

```
text-*    = text styling (alignment, transform, color)
          Examples: .text-left, .text-center, .text-uppercase-tracked, .text-secondary

flex-*    = flexbox properties
          Examples: .flex-center, .flex-between, .flex-col, .flex-gap-md

grid-*    = grid alignment properties
          Examples: .grid-center, .grid-span-2

bg-*      = background properties
          Examples: .bg-primary, .bg-secondary

border-*  = border properties
          Examples: .border-accent, .border-dashed
```

**Why this matters:** `text-right` clearly indicates `text-align: right` (inline content only), while `flex-end` indicates `justify-content: flex-end` (flex children). This prevents confusion about which elements will actually be affected.

---

## 3. `:global()` Usage Rules

### ✅ ALLOWED (Rare)

**Use case:** Styling slot content you don't control

```svelte
<!-- Table.svelte provides structure, consumer provides rows -->
<table>
  <tbody>
    {#each data as item}
      <slot {item} /> <!-- Consumer renders <tr> here -->
    {/each}
  </tbody>
</table>

<style>
  :global(tbody tr) {
    border-bottom: 1px solid var(--ed-border);
  }
  :global(tbody td.align-left) {
    text-align: left;
  }
</style>
```

**Why allowed:** Table owns the structure but consumers provide content. Global styles necessary for consistent table styling.

---

### ⚠️ RARE EXCEPTION

**Use case:** Class forwarding for `style.css` utilities ONLY

```svelte
<!-- Consumer forwards global utility class -->
<Card class="flex-center" />

<!-- Card.svelte accepts and applies it -->
<script>
  let className = '';
  export { className as class };
</script>
<div class="card {className}">
  <slot />
</div>
```

**Constraints:**
- ONLY for utilities defined in `style.css`
- NOT for custom one-off styling (use component props instead)
- Component should accept `class` prop for extension

**When NOT to do this:** If you need custom styling frequently, add a prop instead:

```svelte
<!-- PREFER THIS -->
<Modal width="800px" />

<!-- OVER THIS -->
<Modal class="wide-modal" />
<style>
  :global(.wide-modal) { width: 800px; }
</style>
```

---

### ❌ FORBIDDEN

**1. Parent styling child component internals**

```svelte
<!-- ParentComponent.svelte -->
<ChildComponent />

<style>
  /* DON'T DO THIS - reaching into child */
  :global(.child-internal-class) {
    color: red;
  }

  /* OR THIS - targeting child structure */
  .parent :global(.child-class) {
    color: red;
  }
</style>
```

**Why forbidden:** Tight coupling. Parent shouldn't know about child's internal structure. If you need customization, child should expose props or slots.

**Alternative:**
```svelte
<!-- Child exposes customization via props -->
<ChildComponent primaryColor="red" />
```

---

**2. Cross-component class sharing**

```svelte
<!-- ParentComponent.svelte -->
<style>
  /* DON'T DO THIS - defining global class for child to use */
  :global(.shared-utility) {
    display: flex;
  }
</style>

<!-- ChildComponent.svelte uses it -->
<div class="shared-utility">
```

**Why forbidden:** Hidden dependency. Not obvious where `.shared-utility` is defined. Maintenance nightmare.

**Alternative:** Extract to `style.css` if genuinely shared utility.

---

**3. Styling your own rendered elements**

```svelte
<!-- DON'T DO THIS -->
<tr class="highlight">

<style>
  :global(tr.highlight) {
    background: orange;
  }
</style>
```

**Why forbidden:** You control the element. No need for `:global()`. This is Svelte scope escape without reason.

**Alternative:** Use local scoped styles:
```svelte
<tr class="highlight">

<style>
  .highlight {
    background: orange;
  }
</style>
```

---

### ⚠️ PRAGMATIC EXCEPTION - When No Better Option Exists

**Use case:** Class forwarding for custom styling when props and utilities aren't logical

Sometimes you need to apply specific styles to a child component instance where:
- It's **not a global utility** (too specific, too many properties)
- It's **not worth a component prop** (too one-off, too contextual)
- **No other reasonable option exists**

**Example:**
```svelte
<!-- ExpeditionActive.svelte -->
<IntersectionObserver class="stats-card-container" />

<style>
  :global(.stats-card-container) {
    position: sticky;
    top: 8px;
    z-index: 10;
    transition: all 0.2s ease;
  }
</style>
```

**Why this is acceptable:**
- `sticky` + `top` + `z-index` is too specific for a global utility
- Adding `sticky`, `top`, and `zIndex` props to IntersectionObserver would be architectural bloat
- This specific positioning is unique to this one usage context
- No cleaner solution exists without over-engineering

**Constraints:**
- Must be truly contextual - unique to this parent/child relationship
- Document WHY props or utilities weren't appropriate
- Use sparingly - if you find yourself doing this frequently, reconsider the architecture

---

## 4. Component Style Ownership

**Rule:** Components OWN their classes. Parents NEVER style them.

### ❌ BAD - Parent styling child's classes

```svelte
<!-- ExpeditionEdit.svelte (parent) -->
<style>
  :global(.jump-index) {
    color: var(--ed-text-dim);
  }
</style>

<!-- RouteEditTable.svelte (child) uses it -->
<td class="jump-index">{index}</td>
```

**Problems:**
- Hidden dependency - not clear where `.jump-index` styles come from
- Parent knows too much about child's structure
- Breaks encapsulation

---

### ✅ GOOD - Component owns its styles

```svelte
<!-- RouteEditTable.svelte defines and uses -->
<td class="jump-index">{index}</td>

<style>
  .jump-index {
    color: var(--ed-text-dim);
  }
</style>
```

**OR if used in 3+ components:**

```css
/* style.css */
.jump-index {
  color: var(--ed-text-dim);
  font-variant-numeric: tabular-nums;
}
```

---

## 5. When to Extract to Global

A class should move from component-local to `style.css` when:

1. **Used in 3+ components** - Duplication is a problem
2. **Identical implementation** - Same CSS properties across all uses
3. **General purpose** - Not specific to one feature domain

### Migration Process

1. **Identify duplication** - Same class name and styles in 3+ files
2. **Verify identical** - Check that all implementations are the same
3. **Extract to `style.css`** - Add with clear naming (e.g., `.text-right`, `.flex-center`)
4. **Remove from components** - Delete local definitions
5. **Test** - Verify all components still work

### Don't Extract Prematurely

❌ **Don't extract if:**
- Only used in 1-2 components
- Implementations differ across components
- Specific to one feature (keep in feature component)

✅ **Do extract when:**
- Clear duplication across 3+ components
- General utility pattern emerges
- Maintenance burden outweighs locality benefit

---

## 6. CSS Variables for Theming

**Use CSS variables from `style.css` for all colors and theme values:**

```css
/* ALWAYS use variables */
.card {
  background: var(--ed-bg-secondary);
  border: 1px solid var(--ed-border);
  color: var(--ed-text-primary);
}

/* NEVER hardcode theme colors */
.card {
  background: #0A0A0A;  /* ❌ DON'T */
  color: #E0E0E0;       /* ❌ DON'T */
}
```

**Available theme variables:**
- Colors: `--ed-orange`, `--ed-orange-dim`, `--ed-orange-bright`
- Backgrounds: `--ed-bg-primary`, `--ed-bg-secondary`, `--ed-bg-tertiary`
- Text: `--ed-text-primary`, `--ed-text-secondary`, `--ed-text-dim`
- Borders: `--ed-border`, `--ed-border-accent`
- Status: `--ed-status-planned`, `--ed-status-active`, `--ed-status-completed`, `--ed-status-ended`
- Semantic: `--ed-success`, `--ed-warning`, `--ed-danger`, `--ed-info`

---

## 7. Component Design Patterns

### Modal Component - Size Philosophy

**CRITICAL RULE: Modal NEVER controls its own size. Content controls size.**

The Modal component is designed to fit its content tightly. The modal provides:
- Background overlay styling
- Border, shadow, and theming
- Header/footer structure
- Close button functionality

The modal **deliberately omits**:
- `width` or `max-width`
- `height` or `max-height`
- Any size constraints

**❌ WRONG - Styling modal size:**
```svelte
<!-- ParentComponent.svelte -->
<Modal class="my-modal">
  <div>Content</div>
</Modal>

<style>
  :global(.my-modal) {
    max-width: 700px;  /* ❌ DON'T DO THIS */
  }
</style>
```

**✅ CORRECT - Content controls size:**
```svelte
<!-- ParentComponent.svelte -->
<Modal>
  <div class="my-content">
    Content
  </div>
</Modal>

<style>
  .my-content {
    max-width: 700px;  /* ✅ Content decides size */
  }
</style>
```

**Why this matters:**
- Separation of concerns - Modal handles overlay/chrome, content handles layout
- Predictable behavior - content is always in control
- No fighting between modal constraints and content constraints
- Easier to reason about sizing issues

**This pattern applies to other container components** - if a component is designed to wrap content, it should fit the content unless there's a specific reason (like viewport overflow protection).

---

## Summary

1. **Location:** Component-local by default, extract to `style.css` after 3+ uses
2. **Naming:** Descriptive prefixes (`text-*`, `flex-*`, `bg-*`)
3. **`:global()`:** Avoid except for slot content styling or pragmatic class forwarding when no better option exists
4. **Ownership:** Components own their classes, parents don't style them
5. **Theme:** Always use CSS variables, never hardcode colors
6. **Extraction:** Let utilities prove their worth before going global
7. **Container sizing:** Container components (Modal, etc.) fit content - content controls size, not container

**Key principle:** Prefer component encapsulation over global styles. Extract utilities only when duplication becomes a real maintenance burden (3+ components).
