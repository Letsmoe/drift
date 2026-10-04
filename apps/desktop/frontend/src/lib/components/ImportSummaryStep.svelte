<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '@neoworks-dev/ui';
	import type { ImportResult } from '#lib/api/spotifyImport.ts';
	import type { SpotifyImportFlow } from '#lib/import/spotifyImport.svelte.ts';

	let {
		flow,
		result,
		onclose
	}: { flow: SpotifyImportFlow; result: ImportResult; onclose: () => void } = $props();

	const categoryRows = $derived([
		{ label: 'Liked Songs', count: result.liked },
		{ label: 'Albums', count: result.albums },
		{ label: 'Artists', count: result.artists },
		{ label: 'Podcasts', count: result.shows }
	]);
	const visibleRows = $derived(
		categoryRows.filter((row) => row.count.matched + row.count.unmatched > 0)
	);

	async function openPlaylist(playlistId: string) {
		onclose();
		await goto(`/library/playlist/${encodeURIComponent(playlistId)}`);
	}

	function describeUnmatched(artists: string[], playlist: string): string {
		const parts = [...artists];
		if (playlist !== '') {
			parts.push(`in ${playlist}`);
		}
		return parts.join(' · ');
	}
</script>

<div class="flex flex-col gap-4">
	<p class="text-sm text-muted">Import finished. Items without a confident match were skipped.</p>

	<div class="flex flex-col">
		{#each visibleRows as row (row.label)}
			<div class="flex items-center justify-between px-2 py-1.5 text-sm">
				<span>{row.label}</span>
				<span class="text-muted">
					{row.count.matched} matched{#if row.count.unmatched > 0}, {row.count.unmatched} not found{/if}
				</span>
			</div>
		{/each}
		{#each result.playlists as playlist (playlist.key)}
			<div class="flex items-center justify-between gap-3 px-2 py-1.5 text-sm">
				<span class="min-w-0 flex-1 truncate">{playlist.name}</span>
				<span class="text-muted">
					{playlist.matched} matched{#if playlist.unmatched > 0}, {playlist.unmatched} not found{/if}
				</span>
				<Button size="sm" variant="ghost" onclick={() => openPlaylist(playlist.playlistId)}
					>Open</Button
				>
			</div>
		{/each}
	</div>

	{#if result.unmatched.length > 0}
		<details class="rounded-md bg-raised px-3 py-2">
			<summary class="cursor-pointer text-sm font-semibold">
				{result.unmatched.length} not found on YouTube Music
			</summary>
			<ul class="mt-2 flex max-h-56 flex-col gap-1 overflow-y-auto">
				{#each result.unmatched as item, index (index)}
					<li class="text-sm">
						<span>{item.title}</span>
						<span class="text-dim">{describeUnmatched(item.artists, item.playlist)}</span>
					</li>
				{/each}
			</ul>
		</details>
	{/if}

	<div class="flex justify-end gap-2">
		<Button variant="ghost" onclick={() => flow.restart()}>Import more</Button>
		<Button variant="primary" onclick={onclose}>Done</Button>
	</div>
</div>
