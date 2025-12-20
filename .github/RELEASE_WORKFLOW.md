# Release Workflow Documentation

## Overview

This project uses an automated release workflow combining release-please for version management, GitHub Actions for building cross-platform binaries, and automated Nix flake updates.

## Workflow Architecture

### 1. Release-Please (`release-please.yml`)

**Trigger:** Push to `main` branch

**What it does:**
- Analyzes conventional commits since last release
- Creates/updates a release PR with version bump and changelog
- Uses `RELEASE_PLEASE_TOKEN` (PAT) instead of `GITHUB_TOKEN` to trigger downstream workflows

**Configuration:** `release-please-config.json`
```json
{
  "packages": {
    ".": {
      "release-type": "simple",
      "bump-minor-pre-major": true,
      "initial-version": "0.1.0",
      "pull-request-title-pattern": "chore(release): ${version}",
      "extra-files": [...]
    }
  }
}
```

**Key settings:**
- `bump-minor-pre-major: true` - Allows minor version bumps for features before 1.0.0
- `initial-version: "0.1.0"` - Starts versioning at 0.1.0 instead of 1.0.0
- `pull-request-title-pattern` - Makes PR titles follow conventional commits format
- `extra-files` - Updates version in `wails.json` and `frontend/package.json`

**Why PAT is required:**
- `GITHUB_TOKEN` doesn't trigger other workflows (security feature)
- PAT allows release-please to create tags that trigger the release workflow
- See: https://docs.github.com/en/actions/security-guides/automatic-token-authentication#using-the-github_token-in-a-workflow

### 2. Release Workflow (`release.yml`)

**Trigger:** Push to tags matching `v*`

**Jobs:**

#### a) `build`
- Uses reusable workflow `build-reusable.yml`
- Builds binaries for all platforms
- Uploads artifacts with 1-day retention

#### b) `upload-release-assets`
- Downloads all build artifacts
- Uploads to GitHub release using `softprops/action-gh-release`

#### c) `update-nix-flake`
- Runs after assets are uploaded
- Downloads the webkit2_41 Linux binary
- Calculates SHA256 hash using `nix-prefetch-url`
- Updates `flake.nix` with new version and hash
- Builds and validates the Nix package
- Commits changes back to main with `[skip ci]`

**Nix cache optimization:**
- Uses GitHub Actions cache for Nix store
- Exports/imports Nix store to avoid re-downloading packages
- Significantly speeds up subsequent runs

## Build Matrix

### Reusable Build Workflow (`build-reusable.yml`)

**Matrix strategy:**
```yaml
matrix:
  include:
    - os: ubuntu-22.04
      platform: linux
      arch: amd64
      build-args: ""
      variant: ""
    - os: ubuntu-22.04
      platform: linux
      arch: amd64
      build-args: "-tags webkit2_41"
      variant: "-webkit2_41"
    - os: windows-latest
      platform: windows
      arch: amd64
      build-args: ""
      variant: ""
```

**Why two Linux builds:**
- **Standard build** (`webkit2_40`): Compatible with older systems (Ubuntu 22.04, Debian 11)
- **webkit2_41 build**: Required for modern systems and Nix packaging

**Webkit version compatibility:**
- Ubuntu 22.04 and older ship with libwebkit2gtk-4.0.so.37 (webkit 4.0)
- Ubuntu 24.04+ ship with libwebkit2gtk-4.1.so.0 (webkit 4.1)
- Wails `-tags webkit2_41` flag builds against the 4.1 ABI
- See: https://github.com/wailsapp/wails/issues/3581

**Build artifacts:**
- Linux: `ed-expedition-linux-amd64.tar.gz` and `ed-expedition-linux-amd64-webkit2_41.tar.gz`
- Windows: `ed-expedition-windows-amd64.zip`
- macOS: (commented out) `ed-expedition-darwin-amd64.tar.gz`, `ed-expedition-darwin-arm64.tar.gz`

## Nix Flake Integration

### Package Definition (`flake.nix`)

**Libraries required:**
```nix
libs = with pkgs; [
  pkg-config
  glib           # Required for GLib library
  gtk3           # GTK3 windowing
  webkitgtk_4_1  # Webkit 4.1 for webkit2_41 builds
  gsettings-desktop-schemas  # Desktop integration
];
```

**Binary wrapping:**
- Uses `wrapProgram` to inject library paths at runtime
- Sets `LD_LIBRARY_PATH` for shared library discovery
- Sets `XDG_DATA_DIRS` for gsettings schemas

**Source:**
- Downloads `ed-expedition-linux-amd64-webkit2_41.tar.gz` from GitHub releases
- Hash is auto-updated by release workflow
- Uses SRI hash format (`sha256-...`)

**Why webkit2_41 variant:**
- NixOS uses webkit2gtk_4_1 package
- Binary must be built against matching ABI version
- Standard webkit2_40 build would fail with missing library errors

## Conventional Commits

**Required format:**
```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Common types:**
- `feat:` - New features (triggers minor bump)
- `fix:` - Bug fixes (triggers patch bump)
- `docs:` - Documentation only
- `chore:` - Maintenance tasks
- `ci:` - CI/CD changes
- `refactor:` - Code restructuring
- `test:` - Test additions/changes

**Breaking changes:**
- Add `BREAKING CHANGE:` in commit body or footer
- Or use `!` after type: `feat!: breaking change`
- Triggers major version bump (1.0.0+) or minor (pre-1.0.0)

**Examples:**
```
feat(expedition): add route segment preview
fix(journal): handle missing FSDJump events
docs: update README with build instructions
ci: add webkit2_41 build variant
```

## Release Process

### Manual Steps

1. **Development:**
   - Work on feature branches
   - Use conventional commits
   - Create PR to main

2. **Release:**
   - Merge release-please PR when ready to release
   - This creates a tag (e.g., `v0.1.0`)
   - Tag triggers release workflow automatically

3. **Post-release:**
   - Release workflow builds binaries
   - Uploads to GitHub release
   - Updates flake.nix with new version/hash
   - Nix users can `nix flake update` to get new version

### Automated Steps

```
Push to main
    ↓
release-please analyzes commits
    ↓
Creates/updates release PR
    ↓
[Manual: Merge release PR]
    ↓
Tag created (v0.1.0)
    ↓
Release workflow triggered
    ├─ Build all platforms
    ├─ Upload to GitHub release
    └─ Update flake.nix
         ├─ Download webkit2_41 tarball
         ├─ Calculate hash
         ├─ Update version + hash
         ├─ Test build with Nix
         └─ Commit to main [skip ci]
```

## Troubleshooting

### Release workflow not triggering
- Ensure `RELEASE_PLEASE_TOKEN` secret is set
- Verify it's a PAT with `repo` and `workflow` permissions
- Check that release-please uses the PAT, not `GITHUB_TOKEN`

### Nix build fails with library errors
- Verify binary is webkit2_41 variant
- Check `libs` array includes all required libraries
- Test with: `nix build` then `./result/bin/ed-expedition`

### Wrong version number
- Check `initial-version` in `release-please-config.json`
- Verify `bump-minor-pre-major: true` is set
- Delete `.release-please-manifest.json` to reset (force-push required)

### Build matrix not producing variants
- Verify `matrix.variant` is included in artifact names
- Check `matrix.build-args` is used in build step
- Ensure both matrix entries are present for Linux

## References

- [Release Please Documentation](https://github.com/googleapis/release-please)
- [Conventional Commits Specification](https://www.conventionalcommits.org/)
- [Wails Build System](https://wails.io/docs/reference/cli#build)
- [GitHub Actions Workflow Syntax](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions)
- [Nix Flakes](https://nixos.wiki/wiki/Flakes)
