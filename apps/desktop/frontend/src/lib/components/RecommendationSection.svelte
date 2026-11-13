<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import type { RecommendedItem } from '#lib/api/recommend.ts';
	import IconAction from '#lib/components/IconAction.svelte';
	import Label from '#lib/components/Label.svelte';
	import ScrollRow from '#lib/components/ScrollRow.svelte';
	import SongCard from '#lib/components/SongCard.svelte';
	import SongTile from '#lib/components/SongTile.svelte';
	import TileGrid from '#lib/components/TileGrid.svelte';
	import { appearance } from '#lib/appearance.svelte.ts';
	import { trackKindOf } from '#lib/format.ts';
	import { arrowClockwise } from '#lib/icons.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import type { Recommendations } from '#lib/recommend/recommendations.svelte.ts';
	import { colors, gutter } from '#lib/theme.ts';

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

<View style={{ gap: 8 }}>
	<View style={{ direction: 'row', align: 'center', gap: 8, paddingX: gutter.narrow }}>
		<Label size="heading" weight="bold">{title}</Label>
		{#if onrefresh}
			<IconAction
				icon={arrowClockwise}
				iconSize={16}
				boxSize={32}
				color={appearance.dim}
				hoverColor={colors.text}
				hoverBackground={colors.hover}
				disabled={refreshing}
				onpress={onrefresh}
			/>
		{/if}
	</View>
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
</View>
