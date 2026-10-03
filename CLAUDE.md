# Otter Music

A YouTube Music player. No server: every app talks to YouTube Music directly through a shared Go core.

## Git

You're allowed to use git. Every time you make a big change, commit the current worktree if it's dirty before changing anything, then write your changes, choose a short to-the-point commit message, and push to the remote repository. If you are unsure about what to write in the commit message, ask for help. Write an explanation of the change in the body of the commit message if it is not obvious from the title. Always mention that a commit was made by you and not an actual human.

## Layout

```
core/               Go module github.com/neoworks/otter-music/core: the shared, UI-free core
  music/            domain types (Track, Artist, Album, Thumbnail): the one data shape every app sees
  innertube/        YouTube Music InnerTube client (WEB_REMIX); testdata/ holds trimmed real responses
apps/desktop/       Wails v2 app (Go module …/apps/desktop), binds core to the webview
  frontend/         SvelteKit 3 + Tailwind 4 UI (static SPA, embedded into the binary)
apps/android/       planned: Kotlin + Jetpack Compose + Media3, core via gomobile .aar
apps/ios/           planned: Swift + SwiftUI + AVPlayer, core via gomobile xcframework
```

`go.work` ties the Go modules together; the root `package.json` is the bun workspace.

## Architecture

- **Core is a library, not a backend.** No HTTP server, no GraphQL. Desktop calls it through Wails bindings; mobile will call it through gomobile. Keep all logic (search, queue, stream resolving, caching, recommendations) in `core` so the three UIs only display data.
- **Core API is idiomatic Go.** gomobile can't export slices or most structs; a thin adapter package (e.g. `core/mobile`) will wrap the API for mobile later. Don't bend `core` itself around gomobile.
- **Separate UIs per platform.** Desktop is Svelte, Android is Compose, iOS is SwiftUI. Don't share UI code between them; share data shapes through `core/music`.
- **Domain types carry camelCase JSON tags.** Wails generates the TypeScript models from them (`frontend/src/lib/wailsjs/go/models.ts`).

### YouTube extraction (three layers)

1. **Metadata**: InnerTube JSON (search, browse, next). Implemented in `core/innertube`. Response structs model only the fields the parsers read; encoding/json matches InnerTube's camelCase keys case-insensitively, so they need no tags.
2. **Stream URLs** (not built yet): player request + `sig`/`n` deciphering by running parts of YouTube's `player.js` in goja. Use the per-player cipher configs published by MetrolistGroup/faraday instead of maintaining our own.
3. **PoToken** (not built yet): BotGuard needs a real browser environment. Core will define a token-provider interface that each platform implements with its webview (Wails webview, Android WebView, WKWebView).

zemer-cipher and InnerTubeX are GPL-3.0. Read them for reference; copying their code makes this app GPL.

## Commands

Run from the repo root:

| Command                   | What it does                                                                                                                             |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `bun install`             | install frontend deps                                                                                                                    |
| `bun run tools`           | install golangci-lint v2 into `~/go/bin`                                                                                                 |
| `bun run desktop:dev`     | `wails dev` (hot reload for Go + Vite)                                                                                                   |
| `bun run desktop:build`   | production binary → `apps/desktop/build/bin/otter-music`                                                                                 |
| `bun run desktop:start`   | run the last production build                                                                                                            |
| `bun run desktop:install` | build, then install for the current user (`scripts/install-desktop.ts`: Linux → `~/.local/bin` + desktop entry, macOS → `/Applications`) |
| `bun run bindings`        | regenerate `frontend/src/lib/wailsjs` after changing bound Go methods                                                                    |
| `bun run test`            | Go tests for core                                                                                                                        |
| `bun run lint`            | golangci-lint (core + desktop), prettier, eslint, svelte-check                                                                           |
| `bun run format`          | gofumpt + goimports via `golangci-lint fmt`, then prettier                                                                               |

Linux uses WebKitGTK 4.1, so Wails commands pass `-tags webkit2_41`; macOS/Windows don't need it.

After writing changes, run `bun run lint` and `bun run test` and fix everything they report.

## Frontend (apps/desktop/frontend)

- **SvelteKit 3**: config lives in `vite.config.ts` (`sveltekit({...})`), there is no `svelte.config.js`. `$lib` is gone; use the `#lib/...` subpath import (defined in `package.json` `imports`) **with the file extension**: `#lib/format.ts`, `#lib/components/TrackRow.svelte`.
- Static SPA via adapter-static (`ssr = false`). `bun run build` writes `build/`, which `main.go` embeds; `build/.gitkeep` keeps the embed compiling on a fresh clone.
- Components never import `#lib/wailsjs` directly; go through `#lib/api/` so bindings stay swappable.
- Runes only. State with logic goes in `.svelte.ts` classes (see `#lib/search/songSearch.svelte.ts`).

### Design system

`@neoworks-dev/ui` and `@neoworks-dev/config` are **bun-linked** from `../neoworks.dev/packages/` (run `bun link` there once). Use its tokens (`bg-canvas`, `bg-raised`, `text-dim`, `border-line`, …) and components (`ListRow`, `Button`, `SectionHeader`, `LoadingSpinner`, …) before writing new ones; see `node_modules/@neoworks-dev/ui/COMPONENTS.md`. Generic primitives belong in the ui package (with a Storybook story there); music-specific components go in `frontend/src/lib/components`.

## Code Style

When editing or generating code, prioritize readability and maintainability over cleverness.

Rules:

- Preserve descriptive names. Do not shorten identifiers.
  - Good: `recordId`, `customerAccount`, `paymentMethod`
  - Bad: `rid`, `acct`, `pm`

- Use `camelCase` for variables, functions, parameters, and object fields unless the language, framework, or existing codebase requires another convention.
- Prefer explicit control flow over shorthand.
- Avoid `??`, ternary `?:`, and compact conditional expressions unless they clearly prevent a large amount of repetitive code without reducing readability.
- Do not deeply nest logic.
  - More than 2 indentation levels is too much.
  - Use guard clauses, early returns, helper functions, or extracted validation steps instead.

- Keep functions short and focused.
  - Split large functions into smaller functions instead of writing one huge function.
  - Each function should have one clear responsibility.

- Add comments only when the code is not immediately readable when skimming.
- Comments must be technical, concise, and useful.
  - Good: `// Normalize external IDs before database lookup.`
  - Bad: `// Now we loop through the items and do the thing.`

- Do not add obvious comments that restate the code.
- Prefer clearly named helper functions over long inline logic with comments.
- Do not change behavior, public APIs, data shapes, validation rules, or side effects unless explicitly asked.
- Match the surrounding code style when it conflicts with these rules.

## Tests

Go tests live next to the code (`*_test.go`). InnerTube parsers are tested against trimmed real responses in `testdata/`; when YouTube changes a response shape, capture a fresh one, trim it to the fields the parser reads, and add a test before fixing the parser.
