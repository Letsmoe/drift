<script lang="ts">
	import { Button } from '@neoworks-dev/ui';
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import type { RecommendedItem } from '#lib/api/recommend.ts';
	import ScrollRow from '#lib/components/ScrollRow.svelte';
	import SongCard from '#lib/components/SongCard.svelte';
	import SongTile from '#lib/components/SongTile.svelte';
	import TileGrid from '#lib/components/TileGrid.svelte';
	import { trackKindOf } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import type { Recommendations } from '#lib/recommend/recommendations.svelte.ts';

	let {
		title,
		items,
		context,
		recommendations,
		layout,
		onrefresh,
		refreshing = false
	}: {
		title: string;
		items: RecommendedItem[];
		context: string;
		recommendations: Recommendations;
		layout: 'grid' | 'row';
		onrefresh?: () => void;
		refreshing?: boolean;
	} = $props();

	const tracks = $derived(items.map((item) => item.track));

	function playItem(index: number): void {
		player.playOrToggle(tracks[index], tracks, index, context, trackKindOf(tracks[index]));
	}
</script>

<section class="flex flex-col gap-2">
	<div class="flex items-center gap-2 px-3">
		<h2 class="text-2xl font-bold">{title}</h2>
		{#if onrefresh}
			<span title="Refresh recommendations" class="flex">
				<Button
					round
					size="sm"
					variant="ghost"
					icon={ArrowClockwiseIcon}
					disabled={refreshing}
					onclick={onrefresh}
				/>
			</span>
		{/if}
	</div>
	{#if layout === 'grid'}
		<TileGrid count={items.length}>
			{#snippet tile(index)}
				<SongTile
					track={items[index].track}
					{tracks}
					{index}
					{context}
					onplay={() => recommendations.click(items[index], index)}
					onvisible={() => recommendations.impress(items[index], index)}
				/>
			{/snippet}
		</TileGrid>
	{:else}
		<ScrollRow>
			{#each items as item, index (item.track.videoId)}
				<SongCard
					track={item.track}
					onplay={() => {
						recommendations.click(item, index);
						playItem(index);
					}}
					onvisible={() => recommendations.impress(item, index)}
				/>
			{/each}
		</ScrollRow>
	{/if}
</section>
