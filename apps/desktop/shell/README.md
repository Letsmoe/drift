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

The shell needs GStreamer 1.20 or newer with the base, good and bad plugins: playbin and the Opus decoder come
from base, souphttpsrc from good, and the AAC decoder (`faad`) from bad. `gst-libav` (FFmpeg) is deliberately not
required. `cargo build --release` in this directory builds it; seed `target/` with a
copy of svelte-gpui's own `target/release` first (`cp -r --reflink=auto`) so GPUI is not compiled again.

## Working on pages and components (no Rust needed)

Run these from the repository root. Screenshots and drives use svelte-gpui's offscreen display, so nothing
opens on your desktop.

```
# Build the UI bundle (apps/desktop/frontend/dist/drift.js). Pages in src/pages are picked up by name.
cd apps/desktop/frontend && bun build.mjs            # once
cd apps/desktop/frontend && bun build.mjs --watch    # on every save

# Screenshot: builds the Go backend and the bundle, starts a backend, runs the plain svelte-gpui host.
# Steps given after -- run 0.9 s apart, so SECONDS must cover them. Look at the PNG afterwards.
bun scripts/desktop-shot.ts /tmp/home.png 4
bun scripts/desktop-shot.ts /tmp/search.png 8 -- \
  --do "press:@placeholder=What do you want to play?" --do wait \
  --do "type:@placeholder=What do you want to play?:daft punk" \
  --do "key:@placeholder=What do you want to play?:enter"

# A window you can watch and reload on save (opens on your real display; not for agents).
bun run desktop:dev --plain
```

Each run gets its own backend, bundle and config directory (`/tmp/drift-shot-config-<pid>`), so runs can go side
by side and never touch an installed Drift's library. Set `DRIFT_SHOT_CONFIG` to use a directory you prepared.

A field must have focus before `type:` is used on it (press it first, then `wait`): the host reports the
focus after the typed keys, and the app's Space shortcut would otherwise eat the spaces.

The plain host has no Rust commands, so there is no audio and no `open_external`; everything that
talks to the backend (RPC over WebSocket) works. It gets the backend address from `DRIFT_BACKEND_URL`
and `DRIFT_BACKEND_TOKEN`, which the scripts above set when they build the bundle.

### Something playing

`--track` puts a `music.Track` (JSON) into the backend's queue about four seconds after the window opened
and starts it, as if it had been picked. Repeat it for more tracks.

```
bun scripts/desktop-shot.ts /tmp/queue.png 8 --track '{"videoId":"dQw4w9WgXcQ","title":"Never Gonna Give You Up","artists":[{"id":"UCuAXFkgsw1L7xaCfnd5JJOw","name":"Rick Astley"}],"durationSeconds":213,"thumbnails":[{"url":"https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg","width":480,"height":360}]}'
```

The queue panel, the player bar and the cover tint then show real data. Audio needs `--shell` (below).

`--path /album/<id>` opens the app on that page instead of Home (the path is baked into the bundle as
`DRIFT_START_PATH`), and `WIDTH` and `HEIGHT` in the environment set the window's size (1100x760 by default).

## With the shell (audio)

```
bun run desktop:dev                                   # the shell, window reloads on save
DRIFT_AUDIO_SINK=fakesink bun scripts/desktop-shot.ts /tmp/shot.png 12 --shell --track '{...}'
```

`DRIFT_AUDIO_SINK=fakesink` plays in real time to nowhere, so tests make no sound. The script prints the file
that `DRIFT_AUDIO_LOG` wrote: every audio event as one JSON line (position, duration, state, buffering,
ended, error, the end of a crossfade, and the volume of each slot, with the milliseconds since the engine started).

The crossfade runs in the engine (`audio_crossfade`, see `src/audio/fade.rs`); the UI only says when. The shell is
also an MPRIS player, `org.mpris.MediaPlayer2.drift` (`drift.instance<pid>` when another Drift holds that name), so
`playerctl -p drift status`, `metadata`, `pause` and `next` work on the instance a script started, offscreen too.

## Building and installing

```
bun run desktop:build     # bundle, backend (build/bin/drift-backend), shell (target/release/drift)
bun run desktop:start     # runs that build with --bundle
bun run desktop:install   # Linux: one packaged executable plus the backend in ~/.local/share/drift
```

`desktop:install` runs svelte-gpui's packager on `../frontend` (its `package.json` has the `"svelte-gpui"` key with
the name, id and `crate`), which builds this crate with the `package` feature so the bundle is inside the executable.
The executable and `drift-backend` go to `~/.local/share/drift`, `~/.local/bin/drift` links to the executable, and
`~/.local/share/applications/drift.desktop` is the entry the MPRIS player names as its `DesktopEntry`.

The shell finds the backend in this order: `DRIFT_BACKEND_BINARY`, next to the executable (the installed layout,
through the link as well), under `backend/` next to it, and `../build/bin` of this crate (the development build).

## What the UI can rely on

- `#lib/router.svelte.ts`: `goto(path)`, `back()`, `forward()` and the reactive `router` (`path`,
  `params`, `query`, `page`, `canGoBack`, `canGoForward`). Pages get `{ params, query }` as props.
- `#lib/storage.ts`: `getItem(key)`, `setItem(key, value)` and `removeItem(key)`, synchronous, like `localStorage`.
- `#lib/backend/shell.ts`: `openExternal(url)`, `pickFile`, `pickFolder`.
- `#lib/layout/textFocus.svelte.ts`: every `Input` should call `textFocus.enter()` in `onfocus` and
  `textFocus.leave()` in `onblur`, so the Space shortcut does not play and pause while someone types.
- Pages are `src/pages/<Name>.svelte` with the names in `PageName` of `router.svelte.ts`.
