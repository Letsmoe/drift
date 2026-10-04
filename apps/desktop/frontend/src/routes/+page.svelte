<script lang="ts">
	import { Button, LoadingSpinner } from '@neoworks-dev/ui';
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import RecommendationSection from '#lib/components/RecommendationSection.svelte';
	import ShelfRow from '#lib/components/ShelfRow.svelte';
	import { HomeFeed, type HomeFilter } from '#lib/home/homeFeed.svelte.ts';
	import { Recommendations } from '#lib/recommend/recommendations.svelte.ts';
	import { whenVisible } from '#lib/whenVisible.ts';

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
</script>

{#snippet sectionStatus(status: string, message: string)}
	{#if status === 'loading'}
		<div class="flex justify-center py-8"><LoadingSpinner /></div>
	{:else if status === 'failed'}
		<p class="px-3 text-sm text-dim">Recommendations are unavailable: {message}</p>
	{/if}
{/snippet}

<div class="mb-6 flex gap-2 px-3">
	{#each filters as option (option.value)}
		<button
			type="button"
			aria-pressed={feed.filter === option.value}
			onclick={() => (feed.filter = option.value)}
			class={[
				'duration-fast rounded-full px-4 py-1.5 text-sm font-medium transition-colors',
				feed.filter === option.value && 'bg-action text-action-fg',
				feed.filter !== option.value && 'bg-raised hover:bg-hover'
			]}
		>
			{option.label}
		</button>
	{/each}
</div>

<div class="flex flex-col gap-8">
	{#if showRecommendations}
		{#if recommendations.feedStatus === 'done' && recommendations.feed.length > 0}
			<RecommendationSection
				title="Made for you"
				items={recommendations.feed}
				context="recommend:feed"
				{recommendations}
				layout="grid"
				onrefresh={() => recommendations.refresh()}
			/>
		{:else if recommendations.feedStatus !== 'done'}
			<section class="flex flex-col gap-2">
				<div class="flex items-center gap-2 px-3">
					<h2 class="text-2xl font-bold">Made for you</h2>
					<Button round size="sm" variant="ghost" icon={ArrowClockwiseIcon} disabled />
				</div>
				{@render sectionStatus(recommendations.feedStatus, recommendations.feedError)}
			</section>
		{/if}

		{#if recommendations.rowsStatus === 'done'}
			{#each recommendations.rows as row (row.key)}
				<RecommendationSection
					title={row.title}
					items={row.items}
					context="recommend:{row.key}"
					{recommendations}
					layout="row"
				/>
			{/each}
		{:else}
			{@render sectionStatus(recommendations.rowsStatus, recommendations.rowsError)}
		{/if}
	{/if}

	{#if feed.status === 'loading'}
		<div class="flex justify-center py-24"><LoadingSpinner /></div>
	{:else if feed.status === 'failed'}
		<p class="py-24 text-center text-sm text-red">{feed.errorMessage}</p>
	{:else}
		{#each feed.visibleShelves as shelf, index (index)}
			<ShelfRow {shelf} />
		{/each}
		{#if feed.visibleShelves.length === 0 && !feed.hasMore}
			<p class="py-16 text-center text-sm text-dim">Nothing to show here yet.</p>
		{/if}
	{/if}
</div>

{#if feed.status === 'done' && feed.hasMore}
	<!-- Re-created per page so it fires again when it is still in view after loading. -->
	{#key feed.shelves.length}
		<div class="flex justify-center py-8" {@attach whenVisible(() => feed.loadMore())}>
			<LoadingSpinner />
		</div>
	{/key}
{/if}
