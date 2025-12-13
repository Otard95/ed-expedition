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

**Prerequisites:** Go 1.21+, Node.js, pnpm, and Wails CLI
(Or use `nix develop` if you have Nix installed)

```bash
# Run wails in development mode with hot reload
# It handles installing dependencies and building everything
wails dev
```

> [!NOTE]
> **Linux users:** If you get `webkit2gtk-4.0` pkg-config errors, you likely
> have webkit2gtk 4.1 installed. Use `wails dev -tags webkit2_41` instead.
> The nix flake currently has this issue.
> [Ubuntu 24.04 dependency issue (libwebkit) · Issue #3581 · wailsapp/wails](https://github.com/wailsapp/wails/issues/3581)

### Testing the Journal Watcher

The app monitors Elite Dangerous journal files for real-time tracking. We've built some testing utilities:

- **`cmd/simulate-log`** - Simulates journal file writes with configurable delays
- **`cmd/expected-events`** - Shows what events should be detected from test data
- **`cmd/journal-watcher-test`** - Tests the actual watcher implementation

See `data/` for example journal files to test with.

For detailed architecture, design decisions, and implementation patterns, see:
- `SPEC.md` - Feature specification
- `SPEC_DECISIONS.md` - Design decisions and data structures
- `MODEL_DECISIONS.md` - Python → Go porting considerations

