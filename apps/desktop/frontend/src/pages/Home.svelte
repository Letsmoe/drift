<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { LoadingSpinner } from '@neoworks/svelte-gpui-ui';
	import type { Item, Shelf } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import FilterChip from '#lib/components/FilterChip.svelte';
	import FeaturedPodcast from '#lib/components/FeaturedPodcast.svelte';
	import IconAction from '#lib/components/IconAction.svelte';
	import Label from '#lib/components/Label.svelte';
	import LoadMoreSentinel from '#lib/components/LoadMoreSentinel.svelte';
	import PageList from '#lib/components/PageList.svelte';
	import RecommendationSection from '#lib/components/RecommendationSection.svelte';
	import ShelfRow from '#lib/components/ShelfRow.svelte';
	import { HomeFeed, type HomeFilter } from '#lib/home/homeFeed.svelte.ts';
	import { arrowClockwise } from '#lib/icons.ts';
	import { Recommendations } from '#lib/recommend/recommendations.svelte.ts';
	import { colors, gutter } from '#lib/theme.ts';

	type HomeBlock =
		| { key: string; kind: 'filters' }
		| { key: string; kind: 'madeForYou' }
		| { key: string; kind: 'madeForYouStatus' }
		| { key: string; kind: 'recommendationRow'; rowIndex: number }
		| { key: string; kind: 'rowsStatus' }
		| { key: string; kind: 'feedLoading' }
		| { key: string; kind: 'feedFailed' }
		| { key: string; kind: 'featured'; item: Item }
		| { key: string; kind: 'shelf'; shelf: Shelf }
		| { key: string; kind: 'empty' }
		| { key: string; kind: 'loadMore' };

	const feed = new HomeFeed();
	const recommendations = new Recommendations();
	feed.load();
	recommendations.load();

	const filters: { value: HomeFilter; label: string }[] = [
		{ value: 'all', label: 'All' },
		{ value: 'music', label: 'Music' },
		{ value: 'podcasts', label: 'Podcasts' }
	];
	const showRecommendations = $derived(feed.filter !== 'podcasts');
	const featuredPodcast = $derived(pickFeaturedPodcast());
	const blocks = $derived([...recommendationBlocks(), ...feedBlocks()]);

	// The Podcasts tab opens with one show presented large: the first of the directory.
	function pickFeaturedPodcast(): Item | undefined {
		if (feed.filter !== 'podcasts') {
			return undefined;
		}
		const firstShelf = feed.visibleShelves[0];
		if (!firstShelf) {
			return undefined;
		}
		return firstShelf.items.find((item) => item.kind === 'podcast');
	}

	function recommendationBlocks(): HomeBlock[] {
		const built: HomeBlock[] = [{ key: 'filters', kind: 'filters' }];
		if (!showRecommendations) {
			return built;
		}
		if (recommendations.feedStatus === 'done' && recommendations.feed.length > 0) {
			built.push({ key: 'madeForYou', kind: 'madeForYou' });
		} else if (recommendations.feedStatus !== 'done') {
			built.push({ key: 'madeForYouStatus', kind: 'madeForYouStatus' });
		}
		if (recommendations.rowsStatus !== 'done') {
			built.push({ key: 'rowsStatus', kind: 'rowsStatus' });
			return built;
		}
		recommendations.rows.forEach((row, rowIndex) => {
			built.push({ key: `row:${row.key}`, kind: 'recommendationRow', rowIndex });
		});
		return built;
	}

	function feedBlocks(): HomeBlock[] {
		if (feed.status === 'loading' || feed.status === 'idle') {
			return [{ key: 'feedLoading', kind: 'feedLoading' }];
		}
		if (feed.status === 'failed') {
			return [{ key: 'feedFailed', kind: 'feedFailed' }];
		}
		const built: HomeBlock[] = [];
		if (featuredPodcast) {
			built.push({
				key: `featured:${featuredPodcast.id}`,
				kind: 'featured',
				item: featuredPodcast
			});
		}
		feed.visibleShelves.forEach((shelf, index) => {
			built.push({ key: `shelf:${index}`, kind: 'shelf', shelf });
		});
		if (feed.visibleShelves.length === 0 && !feed.hasMore) {
			built.push({ key: 'empty', kind: 'empty' });
		}
		if (feed.hasMore) {
			built.push({ key: `loadMore:${feed.shelfCount}`, kind: 'loadMore' });
		}
		return built;
	}
</script>

{#snippet sectionStatus(status: string, message: string)}
	{#if status === 'loading'}
		<View style={{ align: 'center', paddingY: 32, color: colors.text }}>
			<LoadingSpinner />
		</View>
	{:else if status === 'failed'}
		<Label size="small" tone="dim" style={{ paddingX: gutter.narrow }}>
			Recommendations are unavailable: {message}
		</Label>
	{/if}
{/snippet}

{#snippet block(item: HomeBlock)}
	{#if item.kind === 'filters'}
		<View style={{ direction: 'row', gap: 8, paddingX: gutter.narrow, marginBottom: -8 }}>
			{#each filters as option (option.value)}
				<FilterChip
					label={option.label}
					selected={feed.filter === option.value}
					onpress={() => feed.select(option.value)}
				/>
			{/each}
		</View>
	{:else if item.kind === 'madeForYou'}
		<RecommendationSection
			title="Made for you"
			items={recommendations.feed}
			context="recommend:feed"
			{recommendations}
			layout="grid"
			onrefresh={() => recommendations.refresh()}
		/>
	{:else if item.kind === 'madeForYouStatus'}
		<View style={{ gap: 8 }}>
			<View style={{ direction: 'row', align: 'center', gap: 8, paddingX: gutter.narrow }}>
				<Label size="heading" weight="bold">Made for you</Label>
				<IconAction
					icon={arrowClockwise}
					iconSize={16}
					boxSize={32}
					color={appearance.dim}
					disabled
				/>
			</View>
			{@render sectionStatus(recommendations.feedStatus, recommendations.feedError)}
		</View>
	{:else if item.kind === 'recommendationRow'}
		{@const row = recommendations.rows[item.rowIndex]}
		{#if row}
			<RecommendationSection
				title={row.title}
				items={row.items}
				context="recommend:{row.key}"
				{recommendations}
				layout="row"
			/>
		{/if}
	{:else if item.kind === 'rowsStatus'}
		{@render sectionStatus(recommendations.rowsStatus, recommendations.rowsError)}
	{:else if item.kind === 'feedLoading'}
		<View style={{ align: 'center', paddingY: 96, color: colors.text }}>
			<LoadingSpinner />
		</View>
	{:else if item.kind === 'feedFailed'}
		<Label size="small" tone="danger" style={{ textAlign: 'center', paddingY: 96 }}>
			{feed.errorMessage}
		</Label>
	{:else if item.kind === 'featured'}
		<View style={{ paddingX: gutter.narrow }}>
			<FeaturedPodcast item={item.item} />
		</View>
	{:else if item.kind === 'shelf'}
		<ShelfRow shelf={item.shelf} />
	{:else if item.kind === 'empty'}
		<Label size="small" tone="dim" style={{ textAlign: 'center', paddingY: 64 }}>
			Nothing to show here yet.
		</Label>
	{:else if item.kind === 'loadMore'}
		<LoadMoreSentinel onvisible={() => feed.loadMore()} />
	{/if}
{/snippet}

<PageList items={blocks} key={(item) => item.key} rowGap={32} estimatedRowHeight={320}>
	{#snippet row(item)}
		{@render block(item)}
	{/snippet}
</PageList>
