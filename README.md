# Elite Dangerous Expedition

ED Expedition is a tool designed to help you plan, execute, and record your
journeys through the galaxy.
It allows you to plot routes using several methods and join one or more
routes into one expedition.

> [!WARNING]
> **Active Development:** This project is in early stages. Expect bugs,
> breaking changes, and missing features. Use at your own risk.

## How does it work?

Created routes using plotters such as Spansh. Every route is simply a series of
jumps from system A to B, based on your ships capabilities.

Expeditions connect multiple routes together at common systems - Sol → Colonia +
Colonia → Beagle Point becomes one seamless journey. The app also allows for
branching paths; Feel like hunting for Earth-like planets mid-trip? Plot a new
route through only F-stars starting and ending on your core route, and ED
Expedition can guide you trough that too.

Once you kick off an journey, the app starts tracking you in-gane jumps. On
every jump, it grabs the next system name on your route, and copies it into
your clipboard so you don't have to tab out.

The app logs each jump - whether you stray on the path, or take a detour. Once
you're finished - or if you halt it yourself - the expedition gets archived
saving the stats and all the jumps, for your records.
<!-- vim-markdown-toc GFM -->

* [Installation](#installation)
    * [Download Pre-built Releases](#download-pre-built-releases)
    * [Running the App](#running-the-app)
* [Contributing](#contributing)
    * [Quick Start](#quick-start)
    * [Commit Messages](#commit-messages)
    * [Configuration](#configuration)
    * [Testing the Journal Watcher](#testing-the-journal-watcher)

<!-- vim-markdown-toc -->
## Installation

### Download Pre-built Releases

**The easiest way to get started:**

1. Go to the [Releases page](https://github.com/Otard95/ed-expedition/releases)
2. Download the latest version for your platform:
   - **Linux:**
     - `ed-expedition-linux-amd64-webkit2_41.tar.gz` for Ubuntu 24.04+ or other recent distros
     - `ed-expedition-linux-amd64.tar.gz` for Ubuntu 22.04/Debian 11 or older systems
     - Not sure? Try webkit2_41 first - if you get library errors, use the standard version
   - **Windows:** `ed-expedition-windows-amd64.zip`
   - **macOS:** Available by request (open an issue)
3. Extract and run the executable

<details>
<summary><strong>Nix / NixOS users</strong></summary>

```bash
# Install to your profile
nix profile add github:Otard95/ed-expedition
```

Or add to your NixOS flake inputs:
```nix
{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";

    ed-expedition = {
      url = "github:Otard95/ed-expedition";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };
}
```

Then add to system packages:
```nix
{ inputs, pkgs, ... }:
{
  environment.systemPackages = [
    inputs.ed-expedition.packages.${stdenv.hostPlatform.system}.default
  ];
}
```
</details>

### Running the App

Just run the binary — the app auto-detects your Elite Dangerous journal directory on first launch. If it can't find it, a dialog will prompt you to select it.

```bash
./ed-expedition        # Linux
ed-expedition.exe      # Windows
```

You can override the journal directory with the `-j` flag or the `ED_EXPEDITION_JOURNAL_DIR` environment variable:

```bash
./ed-expedition -j /path/to/journals
# or
export ED_EXPEDITION_JOURNAL_DIR=/path/to/journals
./ed-expedition
```

The directory is saved after first use — subsequent launches remember it. The `-j` flag acts as a session override if a directory is already saved.

## Contributing

ED Expedition is built with [Wails v2](https://wails.io/docs/introduction) (Go backend + Svelte frontend). If you're new to Wails, check their docs for the framework basics.

### Quick Start

**Prerequisites:** Go 1.23+, Node.js, pnpm, and [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
# Run wails in development mode with hot reload
# It handles installing dependencies and building everything
wails dev
```

> [!TIP]
> **Nix users:** `nix develop` provides all dependencies and automatically sets `ED_DEV_MODE`, `ED_EXPEDITION_DATA_DIR`, `ED_EXPEDITION_CACHE_DIR`, and `ED_EXPEDITION_JOURNAL_DIR` to local project directories.

> [!NOTE]
> **Linux users:** If you get `webkit2gtk-4.0` pkg-config errors, you likely
> have webkit2gtk 4.1 installed (Ubuntu 24.04+, NixOS). Use `wails dev -tags webkit2_41` instead.
> See [wails#3581](https://github.com/wailsapp/wails/issues/3581).

### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/) for all commit messages. This allows automated changelog generation and semantic versioning via [release-please](https://github.com/googleapis/release-please).

**Format:**
```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Common types:**
- `feat:` - New features (triggers minor version bump)
- `fix:` - Bug fixes (triggers patch version bump)
- `docs:` - Documentation changes
- `chore:` - Maintenance tasks (version bumps, dependency updates)
- `ci:` - CI/CD changes
- `refactor:` - Code refactoring
- `perf:` - Performance improvements

**Breaking changes:**
- Use `feat!:` or `fix!:` OR include `BREAKING CHANGE:` in footer
- Triggers major version bump (or minor before v1.0.0)

**Examples:**
```
feat(routes): add Overcharge column with lightning indicator
fix(frontend): correct view button route for completed expeditions
docs: update README with conventional commit guidelines
chore: initialize project version to 0.1.0
```

### Configuration

**Data and Cache Directories:**

By default, expedition/route data is stored in OS-specific locations:
- **Linux:** `~/.local/share/ed-expedition/` (respects `XDG_DATA_HOME`)
- **macOS:** `~/Library/Application Support/ed-expedition/`
- **Windows:** `%APPDATA%\ed-expedition\`

Cache (galaxy database, downloads) uses the OS cache directory:
- **Linux:** `~/.cache/ed-expedition/` (respects `XDG_CACHE_HOME`)
- **macOS:** `~/Library/Caches/ed-expedition/`
- **Windows:** `%LOCALAPPDATA%\ed-expedition\cache\`

Override with environment variables:

```bash
export ED_EXPEDITION_DATA_DIR=/custom/path/to/data
export ED_EXPEDITION_CACHE_DIR=/custom/path/to/cache
export ED_EXPEDITION_JOURNAL_DIR=/path/to/journals
```

**Dev Mode:**

Set `ED_DEV_MODE=1` to enable development behavior (e.g., archive cache files instead of deleting after galaxy build). This is set automatically in the nix dev shell.

### Testing the Journal Watcher

The app monitors Elite Dangerous journal files for real-time tracking. We've built some testing utilities:

- **`cmd/jump-repl`** - Interactive REPL for testing active expeditions. Simulates jumps and targets with live feedback. *Most useful for interactive testing.*
- **`cmd/simulate-log`** - Simulates journal file writes to `./data/journals/` with configurable delays (useful for testing during `wails dev`)
- **`cmd/expected-events`** - Shows what events should be detected from test data
- **`cmd/journal-watcher-test`** - Tests the actual watcher implementation

**Interactive testing with the REPL (recommended):**
```bash
# Terminal 1: Run the app in dev mode
wails dev

# Terminal 2: Start the REPL for your active expedition
cd cmd/jump-repl
go run . ../../data/journals

# In the REPL:
> jump next         # Jump to next expected system
> jump detour       # Jump to a random off-route system
> jump Sol          # Jump to specific system by name
> target next       # Set FSD target without jumping
> status            # Show current expedition state
> help              # Show all commands
```

**Automated testing with simulate-log:**
```bash
# Terminal 1: Run the app in dev mode
wails dev

# Terminal 2: Simulate journal events from a log file
cd cmd/simulate-log
go run . ../../data/test-logs/Journal.2024-10-30T124500.01.log
```
