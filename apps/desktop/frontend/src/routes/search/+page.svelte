<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { LoadingSpinner } from '@neoworks-dev/ui';
	import type { Item, Shelf } from '#lib/api/music.ts';
	import Chip from '#lib/components/Chip.svelte';
	import MediaCard from '#lib/components/MediaCard.svelte';
	import MediaCardRow from '#lib/components/MediaCardRow.svelte';
	import TopResultCard from '#lib/components/TopResultCard.svelte';
	import TrackTable from '#lib/components/TrackTable.svelte';
	import { SearchPageState, shelfTracks } from '#lib/search/searchPage.svelte.ts';

	const topSongCount = 4;

	const results = new SearchPageState();
	const query = $derived(currentQuery());
	const context = $derived(`search:${query}`);

	$effect(() => {
		results.run(query);
	});

	$effect(() => {
		if (query === '') {
			results.loadRecent();
		}
	});

	function currentQuery(): string {
		const urlQuery = page.url.searchParams.get('q');
		if (urlQuery === null) {
			return '';
		}
		return urlQuery.trim();
	}

	function searchFor(recentQuery: string) {
		goto(`/search?q=${encodeURIComponent(recentQuery)}`);
	}

	function itemKindOf(shelf: Shelf): 'song' | 'episode' {
		if (shelf.items[0].kind === 'episode') {
			return 'episode';
		}
		return 'song';
	}

	function hasNoResults(): boolean {
		return results.status === 'done' && results.shelves.length === 0;
	}
</script>

{#snippet trackShelf(shelf: Shelf, visibleCount: number, withHeader: boolean)}
	{@const tracks = shelfTracks(shelf)}
	{#if tracks}
		<TrackTable
			{tracks}
			{context}
			kind={itemKindOf(shelf)}
			{visibleCount}
			showHeader={withHeader}
			showAlbum={withHeader}
		/>
	{/if}
{/snippet}

{#snippet itemGrid(items: Item[])}
	<div class="grid grid-cols-[repeat(auto-fill,11rem)] gap-x-1 gap-y-2">
		{#each items as item (item.id)}
			<MediaCard {item} />
		{/each}
	</div>
{/snippet}

{#if query === ''}
	<section class="flex flex-col gap-4 px-3">
		<div class="flex items-center justify-between">
			<h1 class="text-2xl font-bold">Recent searches</h1>
			{#if results.recentQueries.length > 0}
				<button
					type="button"
					onclick={() => results.clearRecent()}
					class="text-sm font-semibold text-dim hover:text-default"
				>
					Clear all
				</button>
			{/if}
		</div>
		{#if results.recentQueries.length === 0}
			<p class="py-12 text-center text-sm text-dim">Search for songs, albums, artists and more.</p>
		{/if}
		<div class="flex flex-wrap gap-2">
			{#each results.recentQueries as recentQuery (recentQuery)}
				<Chip label={recentQuery} onclick={() => searchFor(recentQuery)} />
			{/each}
		</div>
	</section>
{:else if results.status === 'loading'}
	<div class="flex justify-center py-24"><LoadingSpinner /></div>
{:else if results.status === 'failed'}
	<p class="py-24 text-center text-sm text-red">{results.errorMessage}</p>
{:else if hasNoResults()}
	<p class="py-24 text-center text-sm text-dim">No results found for "{query}".</p>
{:else}
	<div class="flex flex-col gap-6">
		<div class="flex flex-wrap gap-2 px-3">
			<Chip label="All" selected={results.filter === ''} onclick={() => results.select('')} />
			{#each results.filterTitles as title (title)}
				<Chip
					label={title}
					selected={results.filter === title}
					onclick={() => results.select(title)}
				/>
			{/each}
		</div>

		{#if results.filteredShelf}
			{@const shelf = results.filteredShelf}
			<section class="px-3">
				{#if shelfTracks(shelf)}
					{@render trackShelf(shelf, shelf.items.length, true)}
				{:else}
					{@render itemGrid(shelf.items)}
				{/if}
			</section>
		{:else}
			<div class="grid grid-cols-1 gap-6 px-3 lg:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
				{#if results.topResult}
					<section class="flex flex-col gap-2">
						<h2 class="text-2xl font-bold">Top result</h2>
						<TopResultCard item={results.topResult} {context} />
					</section>
				{/if}
				{#if results.songsShelf}
					<section class="flex min-w-0 flex-col gap-2">
						<h2 class="text-2xl font-bold">Songs</h2>
						{@render trackShelf(results.songsShelf, topSongCount, false)}
					</section>
				{/if}
			</div>
			{#each results.otherShelves as shelf (shelf.title)}
				<MediaCardRow title={shelf.title} items={shelf.items} />
			{/each}
		{/if}
	</div>
{/if}
