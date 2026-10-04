<script lang="ts">
	import { Button, Checkbox } from '@neoworks-dev/ui';
	import type { ImportSummary } from '#lib/api/spotifyImport.ts';
	import type { SpotifyImportFlow } from '#lib/import/spotifyImport.svelte.ts';

	let { flow, summary }: { flow: SpotifyImportFlow; summary: ImportSummary } = $props();

	const allPlaylistsSelected = $derived(
		summary.playlists.length > 0 && flow.selectedPlaylistKeys.size === summary.playlists.length
	);
</script>

{#snippet countRow(label: string, count: number, checked: boolean, ontoggle: () => void)}
	<label class="flex cursor-pointer items-center gap-3 rounded-md px-2 py-2 hover:bg-hover">
		<Checkbox {checked} onchange={ontoggle} disabled={count === 0} />
		<span class="flex-1 text-sm">{label}</span>
		<span class="text-sm text-muted">{count}</span>
	</label>
{/snippet}

<div class="flex flex-col gap-3">
	<p class="text-sm text-muted">
		Choose what to bring over. Everything is matched to YouTube Music.
	</p>

	<div class="flex flex-col">
		{@render countRow('Liked Songs', summary.liked, flow.includeLiked, () => {
			flow.includeLiked = !flow.includeLiked;
		})}
		{@render countRow('Albums', summary.albums, flow.includeAlbums, () => {
			flow.includeAlbums = !flow.includeAlbums;
		})}
		{@render countRow('Artists', summary.artists, flow.includeArtists, () => {
			flow.includeArtists = !flow.includeArtists;
		})}
		{@render countRow('Podcasts', summary.shows, flow.includeShows, () => {
			flow.includeShows = !flow.includeShows;
		})}
	</div>

	{#if summary.playlists.length > 0}
		<div class="flex flex-col border-t border-line pt-2">
			<label class="flex cursor-pointer items-center gap-3 rounded-md px-2 py-2 hover:bg-hover">
				<Checkbox checked={allPlaylistsSelected} onchange={() => flow.toggleAllPlaylists()} />
				<span class="flex-1 text-sm font-semibold">Playlists ({summary.playlists.length})</span>
			</label>
			<div class="flex max-h-56 flex-col overflow-y-auto">
				{#each summary.playlists as playlist (playlist.key)}
					<label
						class="flex cursor-pointer items-center gap-3 rounded-md px-2 py-2 pl-8 hover:bg-hover"
					>
						<Checkbox
							checked={flow.selectedPlaylistKeys.has(playlist.key)}
							onchange={() => flow.togglePlaylist(playlist.key)}
						/>
						<span class="min-w-0 flex-1 truncate text-sm">{playlist.name}</span>
						<span class="text-sm text-muted">{playlist.trackCount} tracks</span>
					</label>
				{/each}
			</div>
		</div>
	{/if}

	{#if summary.trackLimit > 0}
		<p class="text-xs text-dim">
			Public links show at most {summary.trackLimit} tracks per playlist; use the data export for the
			full playlist.
		</p>
	{/if}
	{#if flow.cancelled}
		<p class="text-sm text-muted">
			Import stopped. Tracks matched so far are remembered, so starting again continues quickly.
		</p>
	{/if}
	{#if flow.errorMessage !== ''}
		<p class="text-sm text-red">{flow.errorMessage}</p>
	{/if}

	<div class="flex justify-end gap-2 pt-2">
		<Button variant="ghost" onclick={() => flow.restart()}>Back</Button>
		<Button variant="primary" disabled={!flow.canStart} onclick={() => flow.start()}>Import</Button>
	</div>
</div>
