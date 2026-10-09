# Drift

A YouTube Music player for desktop (Electron shell, Go backend), Android (Kotlin) and iOS (Swift), sharing one Go core.

## Setup

```sh
# once: make the neoworks design system linkable
(cd ../neoworks.dev/packages/ui && bun link)
(cd ../neoworks.dev/packages/config && bun link)

bun install
bun run tools   # golangci-lint
bun run desktop:dev       # desktop app with hot reload
bun run desktop:install   # build and install for the current user
```

Requires Go 1.27 and bun; Electron is installed by `bun install`.

See [CLAUDE.md](CLAUDE.md) for layout, architecture and commands.
