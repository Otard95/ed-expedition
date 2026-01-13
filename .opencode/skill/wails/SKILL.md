---
name: wails
description: Wails v2 framework knowledge for Go + web frontend desktop apps. Use when working with Wails projects, debugging build issues, understanding bindings generation, or dealing with platform-specific quirks (webkit versions, live reload, build directory structure).
---

# Wails Framework

Wails is a Go framework for building desktop applications with web frontends. Knowledge here supplements official docs with practical learnings.

## Live Reload Development

When running `wails dev`:

- Frontend changes trigger instant Vite hot reload
- Backend Go changes trigger recompilation and app restart
- TypeScript bindings auto-regenerate when Go exported methods change

**Do not manually run:**
- `pnpm run build`, `go build`, or `wails build`
- `wails generate module` (bindings generate automatically)

**To verify Go compilation without artifacts:** `go build -o /dev/null .`

## TypeScript Bindings

Wails auto-generates TypeScript from exported `App` struct methods:

- Method signatures → `frontend/wailsjs/go/main/App.ts`
- Go structs used in signatures → `frontend/wailsjs/go/models.ts`

All bindings return Promises (async). Backend is source of truth - reload data after mutations.

## Webkit2 Linux Variants

Linux builds must match the system's webkit ABI:

| System | Webkit Package | Wails Flag |
|--------|---------------|------------|
| Ubuntu 22.04, Debian 11 | libwebkit2gtk-4.0 | (none) |
| Ubuntu 24.04+, NixOS | libwebkit2gtk-4.1 | `-tags webkit2_41` |

Building with the wrong variant causes missing library errors at runtime.

## Build Directory Structure

```
build/
├── bin/           # Output binaries
├── darwin/        # macOS-specific (Info.plist, Info.dev.plist)
└── windows/       # Windows-specific (icon.ico, manifest, installer/)
```

Platform files are customizable. Delete them and run `wails build` to regenerate defaults.

## Nix Packaging

For NixOS, the binary needs runtime library wrapping:

```nix
libs = [ glib gtk3 webkitgtk_4_1 gsettings-desktop-schemas ];
```

Use `wrapProgram` to inject `LD_LIBRARY_PATH` and `XDG_DATA_DIRS`.
