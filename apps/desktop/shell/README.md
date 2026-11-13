# Drift desktop shell

The native shell of the desktop app, built on [svelte-gpui](../../../../svelte-gpui): a Rust crate that
starts the Go backend, plays audio with GStreamer, and runs the UI bundle from `../frontend`.

## Setup

The svelte-gpui checkout must sit next to this one (`neoworks/svelte-gpui`). A worktree under
`.worktrees/` reaches it through a symlink: `ln -s <path to svelte-gpui> .worktrees/svelte-gpui`.
Build its host once with `scripts/host.sh` there (`target/release/svelte-gpui`).

```
bun install
```

## Working on pages and components (no Rust needed)

Run these from the repository root. Screenshots and drives use svelte-gpui's offscreen display, so nothing
opens on your desktop.

```
# Build the UI bundle (apps/desktop/frontend/dist/drift.js). Pages in src/pages are picked up by name.
cd apps/desktop/frontend && bun build.mjs            # once
cd apps/desktop/frontend && bun build.mjs --watch    # on every save

# Screenshot: builds the Go backend and the bundle, starts a backend, runs the plain svelte-gpui host.
# Steps given with --do run 0.9 s apart, so SECONDS must cover them. Look at the PNG afterwards.
bun scripts/desktop-shot.ts /tmp/home.png 4
bun scripts/desktop-shot.ts /tmp/search.png 5 -- --do "type:@placeholder=Search:daft punk" --do "key:@placeholder=Search:enter"

# A window you can watch and reload on save (opens on your real display; not for agents).
bun run desktop:dev --plain
```

The plain host has no Rust commands, so there is no audio and no `open_external`; everything that
talks to the backend (RPC over WebSocket) works. It gets the backend address from `DRIFT_BACKEND_URL`
and `DRIFT_BACKEND_TOKEN`, which the scripts above set when they build the bundle.

## With the shell (audio)

```
bun run desktop:dev                                   # the shell, window reloads on save
bun scripts/desktop-shot.ts /tmp/shot.png 5 --shell   # screenshot of the shell
```

## What the UI can rely on

- `#lib/router.svelte.ts`: `goto(path)`, `back()`, `forward()` and the reactive `router` (`path`,
  `params`, `query`, `page`, `canGoBack`, `canGoForward`). Pages get `{ params, query }` as props.
- `#lib/storage.ts`: `getItem(key)`, `setItem(key, value)` and `removeItem(key)`, synchronous, like `localStorage`.
- `#lib/backend/shell.ts`: `openExternal(url)`, `pickFile`, `pickFolder`.
- Pages are `src/pages/<Name>.svelte` with the names in `PageName` of `router.svelte.ts`.
