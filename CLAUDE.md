# Drift

A YouTube Music player. No server: every app talks to YouTube Music directly through a shared Go core. Anonymous by default; an optional YouTube Music sign-in is used only to import the user's library.

## Git

Commit the worktree first if it's dirty, then write your changes. Short, to-the-point commit title; a body explaining the change when the title doesn't carry it; ask if you're unsure what to write. On a feature branch, only commit what that branch is for.

No trailers, ever: no `Co-Authored-By` on a commit, no "Generated with Claude Code" on a PR.

## Writing

Commits and issues carry only what matters. Say the thing, explain what a reader won't see for themselves, stop. No restating the diff, no summarising what you just said, no section that exists because the format seemed to want one.

Write issues and their comments the way you'd explain it to a colleague, in complete sentences.

- Start with a 2–3 sentence summary: what happened, why, and the fix.
- Use short headings, with short paragraphs of normal prose under them.
- Never join ideas with arrows, slashes, colons or dashes. Write "first X, then Y" instead of "X → Y", and "A and B" instead of "A / B".
- Use at most two code identifiers per sentence. Say what each one does the first time it appears.
- Use bullets only for genuinely separate items, and make each bullet a full sentence.
- Use tables only for numbers or timelines.
- Use one date format everywhere: 2026-10-07 18:28.
- Put error messages, paths and commands in code formatting, and long logs in a collapsed `<details>` block.
- Put side findings in a short "Out of scope" section at the end.

## How we work

One agent at a time, with me giving feedback as it goes. Keep each step small enough for me to read in one sitting, and stop to show me rather than piling up work I then have to catch up on.

`main` is what I've reviewed. It stays checked out in the repo root: never switch branches there.

Straight to `main` in the root, no branch: typos, one-liners, and anything that only touches how we work rather than the app — this file, `.claude/skills/`, CI workflows, editor config.

Everything else happens on a branch off `main`, in a worktree under `.worktrees/<branch>` (`git worktree add .worktrees/<branch> -b <branch> main`). Name it `<issue-number>-<slug>` when there is an issue (`12-lyrics-offset`), `<slug>` when there isn't. One branch is one thing; anything found along the way gets written down as an issue and stays off the branch, unless the branch can't finish without it.

Before starting on anything, check whether it is already half-built: `git branch` and `git worktree list` for the feature, and read what is on the branch. Sessions end mid-feature, and a branch is where that work is — starting again writes it a second time and loses whatever the first attempt learned. If a branch for it exists, continue on it.

No pull requests unless I ask for one. Review happens here, on the diff, before the merge.

## Issues

Issues on `letsmoe/drift-music` carry goals across sessions. A session ends and its context goes with it; an issue is the only thing that carries a goal to the next one. So:

- Work that finishes in this session, with me here, needs no issue.
- Work that won't finish in one session gets one, however loosely defined it still is.
- Something concrete found along the way, that isn't what we're doing now, gets one instead of being done.

Write them as soon as the list exists, not once the work starts.

This repo is on my personal account, not neoworks: write to GitHub with plain `gh`, never `gh bot`.

Every issue gets four labels: a type, a priority, an effort and an area. A personal repo has no issue types or issue fields, so all four are labels (`gh issue create --label type:bug --label priority:high --label effort:low --label area:core …`).

- **Type** — `type:bug` something is broken, `type:feature` something new, `type:task` everything else.
- **Priority** — `priority:urgent` drop everything, `priority:high` next up, `priority:medium` normal, `priority:low` nice to have.
- **Effort** — `effort:low` an hour or two, `effort:medium` about a day, `effort:high` several days or needs splitting.
- **Area**, exactly one of:
  - `area:extraction` — `innertube`, `stream`, `lyrics`: talking to YouTube
  - `area:core` — the rest of `core`: library, queue, recommendations, import, downloads, Listen Together
  - `area:desktop` — the Go backend, Electron shell and Svelte frontend
  - `area:android` — the Android app and `core/mobile`
  - `area:release` — CI, packaging, releases

Two areas is fine when an issue genuinely spans them; three means split it.

## Done means verified

Work is done when it has been shown to work, not when the code is written. Shown means one of:

- a test that fails without the fix and passes with it;
- shown in a running instance of the app (the `run` skill), with a screenshot, and one from before if you reproduced it.

`bun run lint` and `bun run test` pass on the branch either way, and `bun run test:live` when `innertube` or `stream` changed.

Then hand it over and stop: what changed in a sentence or two, the evidence, and the branch. When it has an issue, the evidence also goes on the issue as a comment — that comment is what I read, so it's written for someone who wasn't in the session.

Never launch or restart my installed app; run your own instance. When I tell you what's happening, that's the reproduction: take it as given and go find the cause.

## Merging

I review the branch's diff, and it merges into `main` when I say so — never before, and never without evidence. Merge in the root with `git merge --no-ff <branch>`, so the branch stays one unit in the history; if it conflicts, resolve it in the merge commit. Push `main`. Then remove the worktree and delete the branch, and close its issue with `gh issue close <n>`.

## Releases

Publishing a GitHub release (`v<version>`, matching `version` in `apps/desktop/electron/package.json`) runs `.github/workflows/release.yml`, which builds the desktop app for Linux, macOS and Windows and attaches the packages to the release. Release only from reviewed `main`.

## Layout

```
core/               Go module github.com/letsmoe/drift-music/core: the shared, UI-free core
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
  together/         Listen Together: signed identity + friends, Nostr signaling, WebRTC rooms with join approval, queue/playback sync, friends' presence
  mobile/           gomobile adapter: the whole core behind one `Core` object, complex values as JSON strings (Android AAR)
apps/desktop/       desktop backend (Go module …/apps/desktop): core behind a loopback server (SPA, stream proxy, RPC WebSocket)
  electron/         Electron shell: starts the backend, owns the window and native dialogs, packaging
  frontend/         SvelteKit 3 + Tailwind 4 UI (static SPA, embedded into the binary)
apps/android/       planned: Kotlin + Jetpack Compose + Media3, core via gomobile .aar
apps/ios/           planned: Swift + SwiftUI + AVPlayer, core via gomobile xcframework
```

`go.work` ties the Go modules together; the root `package.json` is the bun workspace.

## Architecture

- **Core is a library, not a backend.** No API server, no GraphQL; the only HTTP server is the loopback stream proxy (on desktop it also serves the UI and its RPC socket, see Desktop). Desktop calls core through that RPC; mobile will call it through gomobile. Keep all logic (search, queue, stream resolving, caching, recommendations) in `core` so the three UIs only display data.
- **Core API is idiomatic Go.** gomobile can't export slices or most structs; a thin adapter package (e.g. `core/mobile`) will wrap the API for mobile later. Don't bend `core` itself around gomobile.
- **Separate UIs per platform.** Desktop is Svelte, Android is Compose, iOS is SwiftUI. Don't share UI code between them; share data shapes through `core/music`.
- **Domain types carry camelCase JSON tags.** `bun run bindings` generates the TypeScript models from them (`frontend/src/lib/backend/generated/models.ts`).

### YouTube extraction (three layers)

1. **Metadata**: InnerTube JSON (search, browse, next) via the `WEB_REMIX` profile. The home feed is `browse` `FEmusic_home` plus continuations; continuation responses also carry an empty tab skeleton, so parse `continuationContents` first. Response structs model only the fields the parsers read; encoding/json matches InnerTube's camelCase keys case-insensitively, so they need no tags. Surprises: unfiltered `search` returns one flat list (a `musicCardShelfRenderer` plus one `itemSectionRenderer` per row, rows prefixed "Song • ", no shelf titles), so the parser groups rows by kind into shelves; album/playlist/podcast pages are `twoColumnBrowseResultsRenderer` with the rows in `secondaryContents`, durations in `fixedColumns`, and album rows leave out album, cover and (when equal to the album artist) artists; playlist continuation uses the `continuationItemRenderer` token and answers with `onResponseReceivedActions` (the `continuations` token is a different, wrong shape); podcast episodes are `musicMultiRowListItemRenderer` with spoken durations ("1 hr 5 min") and their continuation response does not name the show, so the `PodcastPage` continuation token wraps the show next to YouTube's token. Video view counts ("922K views") are rewritten as plays (`viewsAsPlays`) in `PlayCount` and card subtitles. Podcast show covers (`i.ytimg.com/pl_c/…/studio_square_thumbnail.jpg?sqp=…`) are signed and answer 404 a few days after they were issued, so `podcasts.Directory.RefreshSavedCovers` renews saved podcasts' covers (desktop: at startup and every 12 h, then the `saved-items` event).
2. **Stream URLs**: `player` requests via the `VISIONOS` profile, which returns plain, unciphered URLs that download fully without a PoToken (yt-dlp's default client without a JS runtime). Every request needs a visitor ID (`visitor_id` endpoint, fetched once per client); without it YouTube answers "Sign in to confirm you're not a bot". Stream URLs are bound to the requesting IP and expire after ~6 h.
3. **PoToken + cipher** (not built, not needed today): `WEB` is SABR-only, `tv` needs a TV session and `web_embedded` rejects music tracks, so there is no ciphered fallback client worth supporting right now. If `VISIONOS` dies, the fallback is yt-dlp's solver ([yt-dlp/ejs](https://github.com/yt-dlp/ejs), Unlicense) plus BotGuard, both run **in the platform webview** (Electron's Chromium, Android WebView, WKWebView) behind an interface in core. Not goja: it could not finish solving one player.js (~3 MB) in 5 minutes.

Profile identities (versions, user agents) mirror yt-dlp's `INNERTUBE_CLIENTS` in `yt_dlp/extractor/youtube/_base.py`; when YouTube rejects a profile, check there first.

**Lyrics** (`core/lyrics`): YouTube first. The lyrics page (`MPLYt…` browse ID from the `next` tabs) is requested with the `ANDROID_MUSIC` profile, which returns per-line timestamps (`timedLyricsModel` → `cueRange.startTimeMilliseconds`); without them the `WEB_REMIX` page gives plain text. Lyrics hang on the audio track: music video IDs usually have none. When YouTube has no synced lyrics and the app allows it (opt-in, off by default; title/artist/album/duration go to a third party), LRCLIB (`lrclib.net/api/get`, then `/api/search` matched by title and ±3 s) is asked; its LRC is parsed by `ParseLRC`.

**Playback path:** `<audio src>` → `StreamURL(videoId)` → loopback proxy `GET /stream/{videoId}` → resolver → CDN. The proxy forwards `Range` for seeking and re-resolves once when the CDN answers 403/410. It always sends a `Range` upstream (`bytes=0-` when the player sends none, as GStreamer does on its first request): without one the CDN throttles to about playback speed and songs take seconds to start. Set `DRIFT_LOG=debug` to log every stream request. Desktop serves it on the backend's `127.0.0.1:<random>` origin; the `<audio>` element uses the absolute URL. Mobile will reuse the same proxy. When a video is downloaded (see Downloads) the proxy serves the local file with `http.ServeContent` instead, so seeking and offline playback work with no player change; downloaded covers are at `GET /cover/{videoId}`.

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
- **Endpoints** (ytmusicapi `mixins/library.py`; parsers reuse `parseCard`/`parseSongs`, fixtures in `testdata/` are hand-synthesized in ytmusicapi's shapes, no real signed-in capture exists, so verify with `DRIFT_TEST_COOKIES=<cookie header> bun run test:live` after touching them): liked songs = browse `VLLM` (playlist continuations), `FEmusic_liked_playlists` (grid; `LM`/`SE` auto playlists skipped), `FEmusic_liked_albums` (grid), `FEmusic_library_corpus_artists` (subscriptions, list rows), `FEmusic_library_non_music_audio_list` (podcasts, grid), `account/account_menu` (name, handle, email, photo). Grids and lists follow `nextContinuationData`, `continuationItemRenderer` and `appendContinuationItemsAction` continuations.
- **Import**: source kind `youtube-account`, exact IDs, no `import_match`. Liked songs are liked oldest first (the newest YouTube like stays the newest local like; already-liked tracks are left alone), playlists mirror into local playlists via `ImportPlaylist` (source `youtube`, key = playlist ID without `VL`), albums/artists/podcasts use `Save`. Re-sync is running it again; `library.MarkYouTubeSynced` records the time.
- **Storage**: migration `003_youtube_session.sql`, one row (`youtube_session`: cookie header, account name/handle/email/photo, created/validated/last-synced). The cookie header is a credential: `Session` and `library.YouTubeSession` redact themselves in logs and formatting, errors never carry it. Sign-out deletes the row.

## Downloads

`core/download` is an in-app offline cache (like YouTube Music Premium), not exported music: the resolver's best audio stream is stored as-is under `<dataDir>/downloads/<videoId>.<webm|m4a>` next to a cover image (`<videoId>.cover.jpg`), `dataDir` being the directory of `library.db`.

- **Queue**: the `download` table (migration 004) is the persistent queue; states `queued`/`downloading`/`done`/`failed` (`paused` is only reported while the queue is paused or blocked). 2 workers, each streams into `<file>.part` with `Range` (always sent, like the proxy), resumes from the partial file, retries 4 times with backoff and re-resolves the URL when the CDN answers 403/410. Interrupted downloads continue at the next start. Progress events are throttled to ~4/s per download in core; the apps only forward them (desktop: RPC event `downloads`, mobile: `DownloadListener`).
- **Collections**: albums, YouTube playlists, local playlists and Liked Songs (`album`, `playlist`, `local`, `likes` with id `liked`). A collection's tracks are recorded in `download_collection_track`; a file is deleted when it is neither downloaded explicitly (single-track download) nor listed by any collection (reference counting). **Keep downloaded** collections sync: local playlist/likes edits via `library.OnChange` (immediately), YouTube albums/playlists at startup, every 6 h and when their page opens (`RefreshCollection`).
- **Offline**: Liked Songs, local playlists, downloaded lists and the queue work for downloaded tracks; `Downloaded()` lists carry the local cover as thumbnail. Catalog pages fail as before; autoplay and recommendation sources log network failures at debug level only.
- **Network policy**: `SetAllowDownloads(false)` (mobile "Download only on Wi-Fi") interrupts like pause; the setting itself lives in the app.
- Tests use an `httptest` fake CDN (Range, 403, cut transfers); `TestLiveDownloadsASearchedTrack` runs with `bun run test:live`.

## Listen Together

`core/together`, ported from otter's Watch Together (`apps/backend/internal/together`, same owner). No server: peers meet on public Nostr relays (`DefaultRelays`, ephemeral kind 25051), then talk peer-to-peer over a WebRTC data channel (pion, STUN only; symmetric NAT on both ends fails, like in otter).

- **Identity**: an Ed25519 key per user (migration 007, `together_identity`; the seed is a credential and redacts itself). The fixed user ID friends add is the first 60 bits of `sha256(publicKey)` as 12 symbols (`K7QD-9MZA-X2PT`). Every relay message is an envelope signed over `topic + body`, and receivers drop anything not signed by the ID it claims (`openSigned`), so nobody can pose as a trusted friend.
- **Topics**: each sealed with AES-GCM under an scrypt key derived from a user ID plus a salt; the topic is a hash of that key. Room = host's ID (`drift-room-v1`), invite inbox = recipient's ID, presence = sender's ID. Anyone who knows an ID can reach these topics; the signatures decide what is believed.
- **Rooms**: star around the host, who forwards between guests. A guest repeats `join` every 2 s; the host offers at once to trusted friends (`together_friend.trusted`) and keeps everyone else in `Requests` (answering `waiting`, guest status `awaiting`, 3 min patience) until the user accepts (optionally trusting them from then on) or declines (`declined`). Every new connection asks again.
- **Sync** (`sync.go`): queue states go around whole with a Lamport version + origin peer ID (ties break by origin, so everyone converges); receivers `queue.Adopt` them, keeping entry IDs so the current track doesn't restart. A guest shares nothing until it adopted the host's queue; only the host's autoplay extends (desktop's extender returns nothing while `Following()`). Play/pause/seek go out as playback messages (tolerance 0.75 s); the host also sends its position every 5 s (tolerance 2 s). New guests get the host's queue and extrapolated position.
- **Presence** (`presence.go`, lobby in `lobby.go`): while online and sharing (`share_presence`, on by default), the current track (smallest cover only), position, paused and whether a session is open go to the presence topic on changes (1 s debounce, seeks > 3 s) and every 20 s; friends are offline after 65 s of silence or an explicit offline message (sent on quit and when sharing is turned off). The lobby runs once the user has friends or used Listen Together; one relay subscription covers the inbox and all friends' presence topics.
- Desktop: `app_together.go`, events `together` (State) and `together-playback`; the player reports to `TogetherReportPlayback` every 5 s and on user actions. Tests run two instances over in-process signaling and loopback WebRTC; `TestLiveListenTogetherOverPublicRelays` uses the real relays (`bun run test:live`).

## Desktop

Electron shell + Go backend, two processes (WebKitGTK was too slow for the UI on Linux; Chromium renders the same design at less than half the CPU).

- **Backend** (`apps/desktop`, binary `drift-backend`): `App` (package main) holds every method the UI calls. `main.go` listens on `127.0.0.1:0`, serves the embedded SPA (`frontend/build`, index.html fallback), the stream proxy (`/stream`, `/cover`) and `/rpc`, then prints `drift-backend {"url":…,"token":…}` on stdout. With `-exit-with-stdin` (set by the shell) it shuts down when stdin closes, so it dies with Electron.
- **Data directory**: `<UserConfigDir>/drift` (`library.db`, `downloads/`). The app was called Otter Music before: `migrateLegacyDirectory` (`app.go`) renames an existing `<UserConfigDir>/otter-music` to it when the new one does not exist.
- **RPC** (`hub.go`, `rpc.go`): one WebSocket, JSON text messages. Request `{id, method, params: [...]}` → `{id, result}` or `{id, error}`; events `{event, data}` (`queue`, `downloads`, `import-progress`/`import-done`/`import-error`) go to every connection through `App.emit`. Any exported `App` method is callable (reflection, positional JSON params; returns nothing, a value, an error or both); `?token=` is required, the origin is not checked. Nothing here is Electron-specific, so a TUI, CLI or remote can be another client.
- **Bindings**: `bun run bindings` = `go run . -bindings …` (`bindings.go`: reflection for types, the Go AST for parameter names) → `frontend/src/lib/backend/generated/{App,models}.ts`, interfaces grouped by Go package (`music.Track`, `main.QueueStep`); pointers are `T | null`, `omitempty` fields optional, `time.Time` a string. Transport: `#lib/backend/rpc.ts` (`call`, `onEvent`).
- **Shell** (`electron/src`): `main.ts` starts the backend, loads `drift://app/` (a protocol handler forwarding to the backend, so the origin and `localStorage` survive the backend's random port; dev loads Vite at `localhost:5173`), sandboxed window, links open in the browser. `preload.ts` exposes `window.drift` (`backend()`, `pickFile`, `pickFolder`; typed in `#lib/backend/shell.ts`). Packaged with electron-builder (config in `electron/package.json`; the backend is an extra resource under `resources/backend/`).

## Commands

Run from the repo root:

| Command                   | What it does                                                                                                                                                                     |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `bun install`             | install frontend deps                                                                                                                                                            |
| `bun run tools`           | install golangci-lint v2 into `~/go/bin`                                                                                                                                         |
| `bun run desktop:dev`     | Vite (hot reload) + fresh backend build + Electron (`scripts/desktop-dev.ts`; restart for Go changes)                                                                            |
| `bun run desktop:build`   | frontend, backend (`build/bin/drift-backend`), Electron bundle, `electron-builder --dir` → `apps/desktop/build/electron/`                                                  |
| `bun run desktop:start`   | run the last production build (Linux)                                                                                                                                            |
| `bun run desktop:install` | build, then install for the current user (`scripts/install-desktop.ts`: Linux → `~/.local/share/drift/app` + `~/.local/bin` link + desktop entry, macOS → `/Applications`) |
| `bun run bindings`        | regenerate `frontend/src/lib/backend/generated` after changing exported `App` methods or their types                                                                             |
| `bun run android:core`    | build `apps/android/app/libs/drift-core.aar` from `core/mobile` (needs gomobile + Android SDK/NDK; skips when up to date)                                                        |
| `bun run android:build`   | debug APK (`gradlew assembleDebug`; builds the core AAR first)                                                                                                                   |
| `bun run android:install` | install the debug build on the connected device/emulator                                                                                                                         |
| `bun run android:start`   | install, then launch the app                                                                                                                                                     |
| `bun run android:logs`    | logcat of the running app only                                                                                                                                                   |
| `bun run android:lint`    | Android lint                                                                                                                                                                     |
| `bun run test`            | Go tests for core                                                                                                                                                                |
| `bun run test:live`       | live tests against real YouTube (`-tags live`), same as the daily canary                                                                                                         |
| `bun run lint`            | golangci-lint (core + desktop), prettier, eslint, svelte-check                                                                                                                   |
| `bun run format`          | gofumpt + goimports via `golangci-lint fmt`, then prettier                                                                                                                       |

After writing changes, run `bun run lint` and `bun run test` and fix everything they report. Run `bun run test:live` after touching `innertube` or `stream`.

## CI

- `.github/workflows/ci.yml`: core (golangci-lint, `go test -race`) and desktop (frontend lint/check, golangci-lint, `go test -race`, `desktop:build`) on push to main and PRs. The desktop job checks out the private `neoworks-dev/neoworks.dev` next to this repo to bun-link the ui package; it needs the `NEOWORKS_DEV_TOKEN` secret (read access to that repo's contents).
- `.github/workflows/canary.yml`: daily live tests against YouTube; on failure it opens an issue labelled `canary`, or comments on the open one. GitHub runner IPs may get bot-checked by YouTube more often than home connections, so a canary failure that doesn't reproduce locally may just be the runner.

## Frontend (apps/desktop/frontend)

- **SvelteKit 3**: config lives in `vite.config.ts` (`sveltekit({...})`), there is no `svelte.config.js`. `$lib` is gone; use the `#lib/...` subpath import (defined in `package.json` `imports`) **with the file extension**: `#lib/format.ts`, `#lib/components/TrackRow.svelte`.
- Static SPA via adapter-static (`ssr = false`). `bun run build` writes `build/`, which `main.go` embeds; `build/.gitkeep` keeps the embed compiling on a fresh clone.
- Components never import `#lib/backend` directly; go through `#lib/api/` so the transport stays swappable.
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
