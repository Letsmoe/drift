<script lang="ts">
	import { FloatingScrollbar } from '@neoworks-dev/ui';
	import type { Item, Shelf, Track } from '#lib/api/music.ts';
	import MediaCard from '#lib/components/MediaCard.svelte';
	import SongTile from '#lib/components/SongTile.svelte';

	let { shelf }: { shelf: Shelf } = $props();

	// Song-only shelves ("Quick picks") render as a compact grid of tiles, the rest as a card row.
	const tracks = $derived(songTracks(shelf.items));

	function songTracks(items: Item[]): Track[] | undefined {
		const tracks: Track[] = [];
		for (const item of items) {
			if (!item.track) {
				return undefined;
			}
			tracks.push(item.track);
		}
		return tracks;
	}
</script>

<section class="flex flex-col gap-2">
	<h2 class="px-3 text-2xl font-bold">{shelf.title}</h2>
	{#if tracks}
		<div class="grid grid-cols-2 gap-2 px-3 xl:grid-cols-3 2xl:grid-cols-4">
			{#each tracks as track, index (track.videoId)}
				<SongTile {track} {tracks} {index} context="home" />
			{/each}
		</div>
	{:else}
		<FloatingScrollbar axis="horizontal">
			<div class="flex">
				{#each shelf.items as item (item.id)}
					<MediaCard {item} />
				{/each}
			</div>
		</FloatingScrollbar>
	{/if}
</section>
