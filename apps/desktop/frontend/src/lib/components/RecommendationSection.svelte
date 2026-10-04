<script lang="ts">
	import { Button, FloatingScrollbar } from '@neoworks-dev/ui';
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import type { RecommendedItem } from '#lib/api/recommend.ts';
	import SongCard from '#lib/components/SongCard.svelte';
	import SongTile from '#lib/components/SongTile.svelte';
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
		player.playOrToggle(tracks[index], tracks, index, context);
	}
</script>

<section class="flex flex-col gap-2">
	<div class="flex items-center gap-2 px-3">
		<h2 class="text-2xl font-bold">{title}</h2>
		{#if onrefresh}
			<Button
				round
				size="sm"
				variant="ghost"
				icon={ArrowClockwiseIcon}
				disabled={refreshing}
				onclick={onrefresh}
			/>
		{/if}
	</div>
	{#if layout === 'grid'}
		<div class="grid grid-cols-2 gap-2 px-3 xl:grid-cols-3 2xl:grid-cols-4">
			{#each items as item, index (item.track.videoId)}
				<SongTile
					track={item.track}
					{tracks}
					{index}
					{context}
					onplay={() => recommendations.click(item, index)}
					onvisible={() => recommendations.impress(item, index)}
				/>
			{/each}
		</div>
	{:else}
		<FloatingScrollbar axis="horizontal">
			<div class="flex">
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
			</div>
		</FloatingScrollbar>
	{/if}
</section>
