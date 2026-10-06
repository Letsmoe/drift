<script lang="ts">
	import type { Item, ItemKind, Shelf, Track } from '#lib/api/music.ts';
	import EpisodeCard from '#lib/components/EpisodeCard.svelte';
	import MediaCard from '#lib/components/MediaCard.svelte';
	import PodcastCard from '#lib/components/PodcastCard.svelte';
	import ScrollRow from '#lib/components/ScrollRow.svelte';
	import SongTile from '#lib/components/SongTile.svelte';
	import TileGrid from '#lib/components/TileGrid.svelte';

	let { shelf }: { shelf: Shelf } = $props();

	// Podcast and episode cards carry descriptions; three per row keeps them readable.
	const threeColumns = 'repeat(3, minmax(0, 1fr))';

	const layout = $derived(layoutOf(shelf.items));
	const tracks = $derived(playableTracks(shelf.items));

	function allOfKind(items: Item[], kind: ItemKind): boolean {
		return items.length > 0 && items.every((item) => item.kind === kind);
	}

	function layoutOf(items: Item[]): 'podcasts' | 'episodes' | 'songs' | 'cards' {
		if (allOfKind(items, 'podcast')) {
			return 'podcasts';
		}
		if (allOfKind(items, 'episode') && items.every((item) => item.track)) {
			return 'episodes';
		}
		if (items.every((item) => item.track)) {
			return 'songs';
		}
		return 'cards';
	}

	function playableTracks(items: Item[]): Track[] {
		const playable: Track[] = [];
		for (const item of items) {
			if (item.track) {
				playable.push(item.track);
			}
		}
		return playable;
	}
</script>

<section class="flex flex-col gap-2">
	<h2 class="px-3 text-2xl font-bold">{shelf.title}</h2>
	{#if layout === 'podcasts'}
		<TileGrid count={shelf.items.length} columns={threeColumns}>
			{#snippet tile(index)}
				<PodcastCard item={shelf.items[index]} />
			{/snippet}
		</TileGrid>
	{:else if layout === 'episodes'}
		<TileGrid count={shelf.items.length} columns={threeColumns}>
			{#snippet tile(index)}
				<EpisodeCard item={shelf.items[index]} track={tracks[index]} {tracks} {index} />
			{/snippet}
		</TileGrid>
	{:else if layout === 'songs'}
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
