<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '@neoworks-dev/ui';
	import type { ImportResult } from '#lib/api/musicImport.ts';
	import type { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';

	let {
		flow,
		result,
		onclose
	}: { flow: MusicImportFlow; result: ImportResult; onclose: () => void } = $props();

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

	const foundLabel = $derived(flow.isAccountSource ? 'imported' : 'matched');
	const missingLabel = $derived(flow.isAccountSource ? 'skipped' : 'not found');

	function describeUnmatched(artists: string[], playlist: string): string {
		const parts = [...artists];
		if (playlist !== '') {
			parts.push(`in ${playlist}`);
		}
		return parts.join(' · ');
	}
</script>

<div class="flex flex-col gap-4">
	<p class="text-sm text-muted">
		{#if flow.isAccountSource}
			Your YouTube Music library is synced. Run it again any time to pick up changes.
		{:else}
			Import finished. Items without a confident match were skipped.
		{/if}
	</p>

	<div class="flex flex-col">
		{#each visibleRows as row (row.label)}
			<div class="flex items-center justify-between px-2 py-1.5 text-sm">
				<span>{row.label}</span>
				<span class="text-muted">
					{row.count.matched}
					{foundLabel}{#if row.count.unmatched > 0}, {row.count.unmatched} {missingLabel}{/if}
				</span>
			</div>
		{/each}
		{#each result.playlists as playlist (playlist.key)}
			<div class="flex items-center justify-between gap-3 px-2 py-1.5 text-sm">
				<span class="min-w-0 flex-1 truncate">{playlist.name}</span>
				<span class="text-muted">
					{playlist.matched}
					{foundLabel}{#if playlist.unmatched > 0}, {playlist.unmatched} {missingLabel}{/if}
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
				{result.unmatched.length}
				{#if flow.isAccountSource}could not be imported{:else}not found on YouTube Music{/if}
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
