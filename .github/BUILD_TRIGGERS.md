# Build Trigger Strategy

## Current Problem

Builds run twice in the release flow:
1. **On push to main** - After merging release-please PR, triggers `build.yml`
2. **On release tag** - Release-please creates tag, triggers `release.yml` → `build-reusable.yml`

This wastes CI minutes and creates duplicate builds.

## Options

### 1. Remove `push` trigger from build.yml (Recommended)

**Change:**
```yaml
# .github/workflows/build.yml
on:
  pull_request:  # Keep - build on PRs for verification
  # Remove: push: branches: [main]
```

**Reasoning:**
- PRs already build and test code before merge
- Release workflow builds on tags for artifacts
- No need to build again on main push (comes from already-tested PR)
- Direct commits to main are rare/discouraged

**Tradeoff:**
- If someone force-pushes to main bypassing PR, no build verification
- Acceptable since PRs are required workflow

### 2. Skip build on release-please commits

**Change:**
```yaml
# .github/workflows/build.yml
on:
  push:
    branches: [main]
  pull_request:

jobs:
  build:
    if: "!contains(github.event.head_commit.message, '[skip ci]')"
    uses: ./.github/workflows/build-reusable.yml
```

**Also configure release-please to use [skip ci]:**
- Not directly supported by release-please
- Would need custom solution

**Reasoning:**
- Release workflow builds anyway
- Release-please commit doesn't need verification

**Tradeoff:**
- Requires release-please to support skip ci (might need workaround)
- More complex conditional logic

### 3. Use concurrency groups to cancel duplicate runs

**Change:**
```yaml
# .github/workflows/build.yml
on:
  push:
    branches: [main]
  pull_request:

concurrency:
  group: build-${{ github.ref }}
  cancel-in-progress: true
```

**Reasoning:**
- If release tag push happens while main push is building, cancel main build
- Reduces wasted resources

**Tradeoff:**
- Doesn't prevent duplicate runs, just cancels them
- Still uses some CI minutes before cancellation

### 4. Only build on PRs and releases (Cleanest)

**Change:**
```yaml
# .github/workflows/build.yml
on:
  pull_request:
  # Remove push trigger entirely

# Rely on release.yml for main branch builds
```

**Reasoning:**
- PR builds verify code quality
- Release builds create artifacts
- Main pushes come from merged PRs (already built)
- No duplicate work

**Tradeoff:**
- No build on direct main commits (acceptable if PRs are enforced)

### 5. Path-based filtering

**Change:**
```yaml
# .github/workflows/build.yml
on:
  push:
    branches: [main]
    paths:
      - 'frontend/**'
      - '**.go'
      - 'go.mod'
      - 'go.sum'
      - 'wails.json'
  pull_request:
    paths:
      - 'frontend/**'
      - '**.go'
      - 'go.mod'
      - 'go.sum'
      - 'wails.json'
```

**Reasoning:**
- Only build when code actually changes
- Docs/CI config changes don't need builds

**Tradeoff:**
- Might miss builds when workflow files change
- Release-please changes multiple files, might still trigger

## Recommendation

**Option 4: Only build on PRs and releases**

Remove `push` trigger from `build.yml`:

```yaml
name: Build

on:
  pull_request:
```

**Why:**
- PRs verify code before merge
- Releases create artifacts
- No wasted duplicate builds
- Simple and predictable

**Alternative: Option 1 + Option 3 (Belt and suspenders)**

Keep push trigger but add concurrency:

```yaml
name: Build

on:
  push:
    branches: [main]
  pull_request:

concurrency:
  group: build-${{ github.ref }}
  cancel-in-progress: true
```

This cancels main build if release tag is pushed immediately after.

## Implementation

To implement Option 4 (recommended):

```bash
# Edit .github/workflows/build.yml
# Remove the push trigger section
```

To implement Option 1 + 3 (alternative):

```bash
# Edit .github/workflows/build.yml
# Add concurrency group
```
