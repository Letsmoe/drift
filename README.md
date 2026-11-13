# Drift

A YouTube Music player for desktop (Rust shell on svelte-gpui, Go backend), Android (Kotlin) and iOS (Swift), sharing one Go core.

## Setup

```sh
bun install
bun run tools   # golangci-lint
bun run desktop:dev       # desktop app with hot reload
bun run desktop:install   # build and install for the current user (Linux)
```

Requires Go 1.27, bun, a Rust toolchain and GStreamer 1.20 or newer. The desktop app also needs a checkout of
[svelte-gpui](https://github.com/neoworks-dev/svelte-gpui) next to this one; see `apps/desktop/shell/README.md`.

See [CLAUDE.md](CLAUDE.md) for layout, architecture and commands.
