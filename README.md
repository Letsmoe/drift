# Otter Music

A YouTube Music player for desktop (Wails), Android (Kotlin) and iOS (Swift), sharing one Go core.

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

Requires Go 1.27, bun, the Wails v2 CLI, and on Linux WebKitGTK 4.1.

See [CLAUDE.md](CLAUDE.md) for layout, architecture and commands.
