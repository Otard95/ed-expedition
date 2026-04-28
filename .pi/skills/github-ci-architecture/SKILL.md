---
name: github-ci-architecture
description: Design decisions and rationale for GitHub Actions CI/CD. Use when modifying workflows, debugging release or build issues, troubleshooting Nix packaging, or understanding why CI is configured the way it is (webkit2_41 variants, PAT tokens, build triggers, release-please).
---

# GitHub CI Architecture

## PAT Token Requirement

Release-please uses `RELEASE_PLEASE_TOKEN` (a PAT) instead of `GITHUB_TOKEN` because GitHub's security model prevents `GITHUB_TOKEN` from triggering downstream workflows. Without the PAT, the release tag creation wouldn't trigger the release build workflow.

Required PAT permissions: `repo` and `workflow`.

## Dual Linux Builds (webkit2_41)

Two Linux binaries are built due to webkit ABI incompatibility:

- **Standard build**: Links against `libwebkit2gtk-4.0` (Ubuntu 22.04, Debian 11)
- **webkit2_41 build**: Links against `libwebkit2gtk-4.1` (Ubuntu 24.04+, NixOS)

Wails flag `-tags webkit2_41` selects the 4.1 ABI. The Nix flake downloads the webkit2_41 variant because NixOS uses `webkitgtk_4_1`.

## Nix Flake Auto-Update

The release workflow automatically updates `flake.nix` after publishing artifacts:

1. Downloads the webkit2_41 Linux tarball
2. Calculates SHA256 hash via `nix-prefetch-url`
3. Updates version and hash in `flake.nix`
4. Commits with `[skip ci]` to avoid retriggering

No manual flake updates needed after releases.

## Build Trigger Strategy

Builds run on **PRs and releases only**, not on push to main.

Rationale: PRs already verify code before merge, and releases build artifacts. Building on main push would duplicate the PR build (since main commits come from merged PRs).

## Conventional Commits

Release-please analyzes commits to determine version bumps:

| Commit Type | Version Bump |
|-------------|--------------|
| `fix:` | Patch (0.0.X) |
| `feat:` | Minor (0.X.0) |
| `feat!:` or `BREAKING CHANGE:` | Minor pre-1.0, Major post-1.0 |

Config: `bump-minor-pre-major: true` allows minor bumps for features before 1.0.0.
