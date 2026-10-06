# Otter Music

A YouTube Music player. No server: every app talks to YouTube Music directly through a shared Go core. Anonymous by default; an optional YouTube Music sign-in is used only to import the user's library.

## Git

You're allowed to use git. Every time you make a big change, commit the current worktree if it's dirty before changing anything, then write your changes, choose a short to-the-point commit message, and push to the remote repository. If you are unsure about what to write in the commit message, ask for help. Write an explanation of the change in the body of the commit message if it is not obvious from the title. Always mention that a commit was made by you and not an actual human.

## Layout

```
core/               Go module github.com/neoworks/otter-music/core: the shared, UI-free core
  music/            domain types (Track, Artist, Album, Thumbnail): the one data shape every app sees
  innertube/        InnerTube client: client profiles, visitor ID, search, home feed, album/playlist/artist/podcast pages, radio, player; `AccountClient` (session-authenticated library reads, the only client that sends cookies); testdata/ holds trimmed real responses (the account/library fixtures are synthesized)
  stream/           resolver (profile chain → best audio URL, cached until expiry) + local HTTP proxy
  library/          local SQLite store (ratings, playlists, saved items, play + search history, impressions, response cache), embedded migrations
  queue/            playback queue: shuffle, repeat, autoplay extension through an Extender; UI-free, apps only render its State
  spotify/          Spotify sources without an account: data-export parser (zip/folder/JSON) + public playlist/album embed-page fetcher; parsing only, no matching
  importer/         library import: Spotify (matches tracks/albums/artists/podcasts to YouTube Music, remembers matches) or the YouTube Music account (exact IDs, no matching); writes likes/playlists/saved items
  account/          optional YouTube Music sign-in: validates + stores the session in the library, hands the importer an authenticated reader
  browsercookies/   desktop-only, UI-free: reads a signed-in browser session (Firefox family, Chromium family) or parses pasted cookies; runs only on the user's click
  recommend/        recommendation engine over the library + a Catalog (the innertube client): candidates → weighted ranking → diversity re-ranking; also the queue's autoplay Extender
  download/         offline downloads: persistent resumable queue, collections kept in sync, local files served by the stream proxy
  mobile/           gomobile adapter: the whole core behind one `Core` object, complex values as JSON strings (Android AAR)
apps/desktop/       Wails v2 app (Go module …/apps/desktop), binds core to the webview
  frontend/         SvelteKit 3 + Tailwind 4 UI (static SPA, embedded into the binary)
apps/android/       planned: Kotlin + Jetpack Compose + Media3, core via gomobile .aar
apps/ios/           planned: Swift + SwiftUI + AVPlayer, core via gomobile xcframework
```

`go.work` ties the Go modules together; the root `package.json` is the bun workspace.

## Architecture

- **Core is a library, not a backend.** No API server, no GraphQL; the only HTTP server is the loopback stream proxy. Desktop calls it through Wails bindings; mobile will call it through gomobile. Keep all logic (search, queue, stream resolving, caching, recommendations) in `core` so the three UIs only display data.
- **Core API is idiomatic Go.** gomobile can't export slices or most structs; a thin adapter package (e.g. `core/mobile`) will wrap the API for mobile later. Don't bend `core` itself around gomobile.
- **Separate UIs per platform.** Desktop is Svelte, Android is Compose, iOS is SwiftUI. Don't share UI code between them; share data shapes through `core/music`.
- **Domain types carry camelCase JSON tags.** Wails generates the TypeScript models from them (`frontend/src/lib/wailsjs/go/models.ts`).

### YouTube extraction (three layers)

1. **Metadata**: InnerTube JSON (search, browse, next) via the `WEB_REMIX` profile. The home feed is `browse` `FEmusic_home` plus continuations; continuation responses also carry an empty tab skeleton, so parse `continuationContents` first. Response structs model only the fields the parsers read; encoding/json matches InnerTube's camelCase keys case-insensitively, so they need no tags. Surprises: unfiltered `search` returns one flat list (a `musicCardShelfRenderer` plus one `itemSectionRenderer` per row, rows prefixed "Song • ", no shelf titles), so the parser groups rows by kind into shelves; album/playlist/podcast pages are `twoColumnBrowseResultsRenderer` with the rows in `secondaryContents`, durations in `fixedColumns`, and album rows leave out album, cover and (when equal to the album artist) artists; playlist continuation uses the `continuationItemRenderer` token and answers with `onResponseReceivedActions` (the `continuations` token is a different, wrong shape); podcast episodes are `musicMultiRowListItemRenderer` with spoken durations ("1 hr 5 min") and their continuation response does not name the show, so the `PodcastPage` continuation token wraps the show next to YouTube's token.
2. **Stream URLs**: `player` requests via the `VISIONOS` profile, which returns plain, unciphered URLs that download fully without a PoToken (yt-dlp's default client without a JS runtime). Every request needs a visitor ID (`visitor_id` endpoint, fetched once per client); without it YouTube answers "Sign in to confirm you're not a bot". Stream URLs are bound to the requesting IP and expire after ~6 h.
3. **PoToken + cipher** (not built, not needed today): `WEB` is SABR-only, `tv` needs a TV session and `web_embedded` rejects music tracks, so there is no ciphered fallback client worth supporting right now. If `VISIONOS` dies, the fallback is yt-dlp's solver ([yt-dlp/ejs](https://github.com/yt-dlp/ejs), Unlicense) plus BotGuard, both run **in the platform webview** (Wails WebKit, Android WebView, WKWebView) behind an interface in core. Not goja: it could not finish solving one player.js (~3 MB) in 5 minutes.

Profile identities (versions, user agents) mirror yt-dlp's `INNERTUBE_CLIENTS` in `yt_dlp/extractor/youtube/_base.py`; when YouTube rejects a profile, check there first.

**Lyrics** (`core/lyrics`): YouTube first. The lyrics page (`MPLYt…` browse ID from the `next` tabs) is requested with the `ANDROID_MUSIC` profile, which returns per-line timestamps (`timedLyricsModel` → `cueRange.startTimeMilliseconds`); without them the `WEB_REMIX` page gives plain text. Lyrics hang on the audio track: music video IDs usually have none. When YouTube has no synced lyrics and the app allows it (opt-in, off by default; title/artist/album/duration go to a third party), LRCLIB (`lrclib.net/api/get`, then `/api/search` matched by title and ±3 s) is asked; its LRC is parsed by `ParseLRC`.

**Playback path:** `<audio src>` → `StreamURL(videoId)` → loopback proxy `GET /stream/{videoId}` → resolver → CDN. The proxy forwards `Range` for seeking and re-resolves once when the CDN answers 403/410. It always sends a `Range` upstream (`bytes=0-` when the player sends none, as GStreamer does on its first request): without one the CDN throttles to about playback speed and songs take seconds to start. Set `OTTER_MUSIC_LOG=debug` to log every stream request. Desktop serves it on `127.0.0.1:<random>` instead of the Wails asset server because WebKitGTK's media pipeline does not seek reliably over custom URI schemes. Mobile will reuse the same proxy. When a video is downloaded (see Downloads) the proxy serves the local file with `http.ServeContent` instead, so seeking and offline playback work with no player change; downloaded covers are at `GET /cover/{videoId}`.

Don't copy code from zemer-cipher or InnerTubeX (GPL-3.0, would make this app GPL) or MetrolistGroup/faraday (no license). Read them for reference only.

## Recommendations

`core/recommend` runs candidate generation → ranking → re-ranking per request; no ML, no account. Every tunable is in `weights.go`.

- **Signals** (per track, decay half-life 30 days, 6-half-life window, each play row its own signal): like 3, completed play (flag or ≥80%) 2, added to a playlist 1.5, partial play 1, clicked impression 0.5; negative: ignored impression −0.15 (cap 3), early skip −1.5 (stopped before 30 s and 50%, plays older than 5 min), dislike −3 (also excluded from all output, with no time window). Followed artists and saved albums add a fixed boost to their affinity.
- **Profile**: computed on demand, never stored. A track's weight goes fully to its first linked artist (÷√n to featured ones) and its album; each dimension is normalized to [−1, 1]. Seeds = strongest tracks with weight ≥ 1, top 8.
- **Candidate sources** (each item tagged with one, prior in `weights.go`): `related:radio` (radio of each seed), `affinity:artist` (top songs of the top 5 artists), `explore:artist` (top songs of up to 4 related artists the profile doesn't know), `rediscover` (liked/completed tracks unplayed for 30 days), `search` (last 3 searches), `youtube` (home feed songs; the cold-start source). Catalog calls run 4 at a time with a timeout, responses are cached in the library (radio/artist/search 24 h, home 1 h), and a failing source is logged and skipped.
- **Ranking**: `affinity + 0.6·prior − penalties` (played in the last day/week, fatigue from ignored impressions); disliked tracks are excluded. **Re-ranking** per 24-item page: ≤2 per artist, ≤3 per album, no source more than twice in a row, every 7th slot exploration; constraints drop weakest first.
- `Feed` memoizes the ranking (first page 10 min, later pages 30 min, `refresh` forces a re-rank); `Rows` groups by reason ("Because you listened to X", "More from <artist>") and drops thin and youtube-only rows. `Extend` (queue autoplay) takes the radio of the last played track, drops disliked/queued/played tracks and ranks by taste.
- **Magic shuffle** (`queue.SetMagicShuffle`, shuffle button cycles off → shuffle → magic): `Mix` merges the radios of up to 5 random user-queued tracks, ranked like `Extend`; the queue puts one recommendation (`Entry.Magic`) after every 3 upcoming user entries (≤30 per mix), mixes again when the user queues more, and removes the upcoming ones when magic shuffle or shuffle is turned off.

## Spotify import

`core/spotify` reads, `core/importer` matches and writes; no Spotify account or API is used (new Spotify developer apps are allowlist-only). Two sources:

- **Data export** (`my_spotify_data.zip`, an extracted folder or a single JSON): `YourLibrary.json` (liked `tracks`, `albums`, `artists`, `shows`) and `PlaylistN.json`. Parsed defensively (unknown/missing fields tolerated); the export has no durations and one artist string per track. Only `YourLibrary*`/`Playlist*` files are read.
- **Public link** (`open.spotify.com/{playlist,album}/<id>`, `spotify:` URIs): the embed page `/embed/<kind>/<id>` carries `<script id="__NEXT_DATA__">` with `props.pageProps.state.data.entity` (`name`, `subtitle`, `trackList` [{uri, title, subtitle "A, B", duration ms}]). It needs a browser User-Agent and is **capped at 100 tracks** per playlist (the UI says so). An album link imports as one saved album.
- **Matching**: query `<title> <first artist>` (`SearchSongs`; albums/artists/podcasts via `Search` and the matching shelf kind). `scoreSong` (pure, in `score.go`) = 0.5 title similarity + 0.3 artist overlap (primary artist weighs 0.7) + 0.2 duration (full within 3 s, 0 beyond 15 s; left out when unknown); feat./remaster/edition noise is stripped; a live/remix/acoustic/... mismatch scales by 0.8, a duration >15 s off by 0.85. Thresholds: song 0.75, album 0.75, artist 0.88 (name only), podcast 0.8. Catalog calls run 4 at a time with a 150 ms pause; 8 failures in a row (or only failures) abort the import.
- **Stored matches**: `import_match` (source, source_uri) -> target id, score and the matched track/item as JSON; "no match" is stored too, so a re-import never searches again (`Options.RetryUnmatched` overrides). `import_playlist` links a Spotify playlist key to its local playlist: a re-import replaces that playlist's entries instead of duplicating it. Likes use `SetRating` (no fake play rows; already-liked tracks are left alone), albums/artists/podcasts use `Save`.
- **API**: `importer.New(lib, catalog, logger)`, `Preview(source)`, `Run(ctx, source, Options, onProgress)`, `Start`/`Cancel` for background runs (one import at a time). Desktop: `app_import.go` + `import-progress`/`import-done`/`import-error` events (shared with the YouTube account import); mobile: `core/mobile/importer.go` (path-based, JSON + `ImportListener`).

## YouTube Music account

Optional sign-in whose only job is reading the user's own library for import and re-sync. Catalog browsing, search, recommendations and playback stay anonymous: only `innertube.AccountClient` attaches the session, never the anonymous `Client` (so the account never touches stream fetching).

- **Cookies**: the `Cookie` header of music.youtube.com (needs `SAPISID` or `__Secure-3PAPISID`). Sources: desktop reads a signed-in browser (`core/browsercookies`) or the user pastes a header, request headers or a "Copy as cURL" command (`ParseCookieHeader`); Android will pass its WebView `CookieManager` cookies to `Core.YouTubeSignIn`. Sign-in validates with `account/account_menu` and stores the session only when that works.
- **Browser support** (implemented): Firefox, LibreWolf, Zen, Floorp, Waterfox on Linux/macOS/Windows (`cookies.sqlite` is copied with its `-wal` to a temp dir, read, deleted; only cookies a browser sends to music.youtube.com, unexpired, container tabs left out). Chrome, Chromium, Brave, Edge, Vivaldi on Linux (v10 "peanuts" key; v11 key from the Secret Service over D-Bus, KWallet through `kwallet-query` as fallback) and macOS (Keychain via `security`, best effort, untested). Chromium browsers on Windows are skipped (app-bound encryption); a profile that cannot be read is listed with a reason instead of failing the rest. Detection only runs from a UI click.
- **Auth**: `Authorization: SAPISIDHASH <ts>_<sha1("<ts> <SAPISID> https://music.youtube.com")>` plus `SAPISID1PHASH`/`SAPISID3PHASH` (built from `__Secure-1PAPISID`/`__Secure-3PAPISID`) space-separated, `Cookie`, `Origin`/`X-Origin: https://music.youtube.com`, `X-Goog-AuthUser: 0`, visitor ID. Signed-out answers (401/403, `logged_in` = 0 in `serviceTrackingParams`, no account header) become `innertube.ErrSessionExpired`.
- **Endpoints** (ytmusicapi `mixins/library.py`; parsers reuse `parseCard`/`parseSongs`, fixtures in `testdata/` are hand-synthesized in ytmusicapi's shapes, no real signed-in capture exists, so verify with `OTTER_MUSIC_TEST_COOKIES=<cookie header> bun run test:live` after touching them): liked songs = browse `VLLM` (playlist continuations), `FEmusic_liked_playlists` (grid; `LM`/`SE` auto playlists skipped), `FEmusic_liked_albums` (grid), `FEmusic_library_corpus_artists` (subscriptions, list rows), `FEmusic_library_non_music_audio_list` (podcasts, grid), `account/account_menu` (name, handle, email, photo). Grids and lists follow `nextContinuationData`, `continuationItemRenderer` and `appendContinuationItemsAction` continuations.
- **Import**: source kind `youtube-account`, exact IDs, no `import_match`. Liked songs are liked oldest first (the newest YouTube like stays the newest local like; already-liked tracks are left alone), playlists mirror into local playlists via `ImportPlaylist` (source `youtube`, key = playlist ID without `VL`), albums/artists/podcasts use `Save`. Re-sync is running it again; `library.MarkYouTubeSynced` records the time.
- **Storage**: migration `003_youtube_session.sql`, one row (`youtube_session`: cookie header, account name/handle/email/photo, created/validated/last-synced). The cookie header is a credential: `Session` and `library.YouTubeSession` redact themselves in logs and formatting, errors never carry it. Sign-out deletes the row.

## Downloads

`core/download` is an in-app offline cache (like YouTube Music Premium), not exported music: the resolver's best audio stream is stored as-is under `<dataDir>/downloads/<videoId>.<webm|m4a>` next to a cover image (`<videoId>.cover.jpg`), `dataDir` being the directory of `library.db`.

- **Queue**: the `download` table (migration 004) is the persistent queue; states `queued`/`downloading`/`done`/`failed` (`paused` is only reported while the queue is paused or blocked). 2 workers, each streams into `<file>.part` with `Range` (always sent, like the proxy), resumes from the partial file, retries 4 times with backoff and re-resolves the URL when the CDN answers 403/410. Interrupted downloads continue at the next start. Progress events are throttled to ~4/s per download in core; the apps only forward them (desktop: Wails event `downloads`, mobile: `DownloadListener`).
- **Collections**: albums, YouTube playlists, local playlists and Liked Songs (`album`, `playlist`, `local`, `likes` with id `liked`). A collection's tracks are recorded in `download_collection_track`; a file is deleted when it is neither downloaded explicitly (single-track download) nor listed by any collection (reference counting). **Keep downloaded** collections sync: local playlist/likes edits via `library.OnChange` (immediately), YouTube albums/playlists at startup, every 6 h and when their page opens (`RefreshCollection`).
- **Offline**: Liked Songs, local playlists, downloaded lists and the queue work for downloaded tracks; `Downloaded()` lists carry the local cover as thumbnail. Catalog pages fail as before; autoplay and recommendation sources log network failures at debug level only.
- **Network policy**: `SetAllowDownloads(false)` (mobile "Download only on Wi-Fi") interrupts like pause; the setting itself lives in the app.
- Tests use an `httptest` fake CDN (Range, 403, cut transfers); `TestLiveDownloadsASearchedTrack` runs with `bun run test:live`.

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
| `bun run android:core`    | build `apps/android/app/libs/otter-core.aar` from `core/mobile` (needs gomobile + Android SDK/NDK; skips when up to date)                |
| `bun run android:build`   | debug APK (`gradlew assembleDebug`; builds the core AAR first)                                                                           |
| `bun run android:install` | install the debug build on the connected device/emulator                                                                                 |
| `bun run android:start`   | install, then launch the app                                                                                                             |
| `bun run android:logs`    | logcat of the running app only                                                                                                           |
| `bun run android:lint`    | Android lint                                                                                                                             |
| `bun run test`            | Go tests for core                                                                                                                        |
| `bun run test:live`       | live tests against real YouTube (`-tags live`), same as the daily canary                                                                 |
| `bun run lint`            | golangci-lint (core + desktop), prettier, eslint, svelte-check                                                                           |
| `bun run format`          | gofumpt + goimports via `golangci-lint fmt`, then prettier                                                                               |

Linux uses WebKitGTK 4.1, so Wails commands pass `-tags webkit2_41`; macOS/Windows don't need it.

After writing changes, run `bun run lint` and `bun run test` and fix everything they report. Run `bun run test:live` after touching `innertube` or `stream`.

## CI

- `.github/workflows/ci.yml`: core (golangci-lint, `go test -race`) and desktop (frontend lint/check, golangci-lint, `wails build`) on push to main and PRs. The desktop job checks out the private `neoworks-dev/neoworks.dev` next to this repo to bun-link the ui package; it needs the `NEOWORKS_DEV_TOKEN` secret (read access to that repo's contents).
- `.github/workflows/canary.yml`: daily live tests against YouTube; on failure it opens an issue labelled `canary`, or comments on the open one. GitHub runner IPs may get bot-checked by YouTube more often than home connections, so a canary failure that doesn't reproduce locally may just be the runner.

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
