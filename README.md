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

## Contributing

ED Expedition is built with [Wails v2](https://wails.io/docs/introduction) (Go backend + Svelte frontend). If you're new to Wails, check their docs for the framework basics.

### Quick Start

**Prerequisites:** Go 1.23+, Node.js, pnpm, and Wails CLI
(Or use `nix develop` if you have Nix installed)

```bash
# Run wails in development mode with hot reload
# It handles installing dependencies and building everything
wails dev
```

> [!WARNING]
> **Journal directory configuration:** The `-j` flag doesn't work with `wails dev` due to how Wails passes flags. The app defaults to `./data/journals` (see `main.go:24`). To test with real journal files, either:
> - Copy/symlink your journal files to `./data/journals/`
> - Use `cmd/simulate-log` to generate test journal events in `./data/journals/`
> - Build the binary (`wails build`) and run it with `-j /path/to/journal`

> [!NOTE]
> **Linux users:** If you get `webkit2gtk-4.0` pkg-config errors, you likely
> have webkit2gtk 4.1 installed. Use `wails dev -tags webkit2_41` instead.
> The nix flake currently has this issue.
> [Ubuntu 24.04 dependency issue (libwebkit) · Issue #3581 · wailsapp/wails](https://github.com/wailsapp/wails/issues/3581)

### Configuration

**Data Directory:**

By default, expedition/route data is stored in OS-specific locations:
- **Linux:** `~/.local/share/ed-expedition/` (respects `XDG_DATA_HOME`)
- **macOS:** `~/Library/Application Support/ed-expedition/`
- **Windows:** `%APPDATA%\ed-expedition\`

Override with the `ED_EXPEDITION_DATA_DIR` environment variable:

```bash
export ED_EXPEDITION_DATA_DIR=/custom/path/to/data
wails dev -j /path/to/journal
```

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
