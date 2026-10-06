<script lang="ts">
	import type { Item, Shelf, Track } from '#lib/api/music.ts';
	import MediaCard from '#lib/components/MediaCard.svelte';
	import ScrollRow from '#lib/components/ScrollRow.svelte';
	import SongTile from '#lib/components/SongTile.svelte';
	import TileGrid from '#lib/components/TileGrid.svelte';

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
		<TileGrid count={tracks.length}>
			{#snippet tile(index)}
				<SongTile track={tracks[index]} {tracks} {index} context="home" />
			{/snippet}
		</TileGrid>
	{:else}
		<ScrollRow>
			{#each shelf.items as item (item.id)}
				<MediaCard {item} />
			{/each}
		</ScrollRow>
	{/if}
</section>
