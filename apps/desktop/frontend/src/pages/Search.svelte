<script lang="ts">
	import { Button, View } from '@neoworks/svelte-gpui/ui';
	import { Chip, LoadingSpinner } from '@neoworks/svelte-gpui-ui';
	import type { Item, Shelf } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import FilterChip from '#lib/components/FilterChip.svelte';
	import Label from '#lib/components/Label.svelte';
	import MediaCard from '#lib/components/MediaCard.svelte';
	import MediaCardRow from '#lib/components/MediaCardRow.svelte';
	import PageList from '#lib/components/PageList.svelte';
	import TopResultCard from '#lib/components/TopResultCard.svelte';
	import TrackTable from '#lib/components/TrackTable.svelte';
	import type { PageProps } from '#lib/pages/pageProps.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { SearchPageState, shelfTracks } from '#lib/search/searchPage.svelte.ts';
	import { colors, gutter } from '#lib/theme.ts';

	type SearchBlock =
		| { key: string; kind: 'recent' }
		| { key: string; kind: 'loading' }
		| { key: string; kind: 'failed' }
		| { key: string; kind: 'noResults' }
		| { key: string; kind: 'filters' }
		| { key: string; kind: 'filteredShelf'; shelf: Shelf }
		| { key: string; kind: 'topSection' }
		| { key: string; kind: 'shelf'; shelf: Shelf };

	const topSongCount = 4;

	let { query: queryParameters }: PageProps = $props();

	const results = new SearchPageState();
	const query = $derived(currentQuery());
	const context = $derived(`search:${query}`);
	const blocks = $derived(buildBlocks());

	$effect(() => {
		results.run(query);
	});

	$effect(() => {
		if (query === '') {
			results.loadRecent();
		}
	});

	function currentQuery(): string {
		const urlQuery = queryParameters.q;
		if (urlQuery === undefined) {
			return '';
		}
		return urlQuery.trim();
	}

	function searchFor(recentQuery: string): void {
		goto(`/search?q=${encodeURIComponent(recentQuery)}`);
	}

	function itemKindOf(shelf: Shelf): 'song' | 'episode' {
		if (shelf.items[0].kind === 'episode') {
			return 'episode';
		}
		return 'song';
	}

	function buildBlocks(): SearchBlock[] {
		if (query === '') {
			return [{ key: 'recent', kind: 'recent' }];
		}
		if (results.status === 'loading') {
			return [{ key: 'loading', kind: 'loading' }];
		}
		if (results.status === 'failed') {
			return [{ key: 'failed', kind: 'failed' }];
		}
		if (results.status === 'done' && results.shelves.length === 0) {
			return [{ key: 'noResults', kind: 'noResults' }];
		}
		return [{ key: 'filters', kind: 'filters' }, ...resultBlocks()];
	}

	function resultBlocks(): SearchBlock[] {
		const filtered = results.filteredShelf;
		if (filtered) {
			return [{ key: `filtered:${filtered.title}`, kind: 'filteredShelf', shelf: filtered }];
		}
		const built: SearchBlock[] = [{ key: 'top', kind: 'topSection' }];
		for (const shelf of results.otherShelves) {
			built.push({ key: `shelf:${shelf.title}`, kind: 'shelf', shelf });
		}
		return built;
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
	<View style={{ direction: 'row', wrap: true, rowGap: 8 }}>
		{#each items as item (item.id)}
			<MediaCard {item} />
		{/each}
	</View>
{/snippet}

{#snippet block(item: SearchBlock)}
	{#if item.kind === 'recent'}
		<View style={{ gap: 16, paddingX: gutter.narrow }}>
			<View style={{ direction: 'row', align: 'center', justify: 'between' }}>
				<Label size="heading" weight="bold">Recent searches</Label>
				{#if results.recentQueries.length > 0}
					<Button
						style={{ color: appearance.dim }}
						hover={{ color: colors.text }}
						onpress={() => results.clearRecent()}
					>
						<Label size="small" weight="semibold">Clear all</Label>
					</Button>
				{/if}
			</View>
			{#if results.recentQueries.length === 0}
				<Label size="small" tone="dim" style={{ textAlign: 'center', paddingY: 48 }}>
					Search for songs, albums, artists and more.
				</Label>
			{/if}
			<View style={{ direction: 'row', wrap: true, gap: 8 }}>
				{#each results.recentQueries as recentQuery (recentQuery)}
					<Chip variant="suggestion" label={recentQuery} onpress={() => searchFor(recentQuery)} />
				{/each}
			</View>
		</View>
	{:else if item.kind === 'loading'}
		<View style={{ align: 'center', paddingY: 96, color: colors.text }}>
			<LoadingSpinner />
		</View>
	{:else if item.kind === 'failed'}
		<Label size="small" tone="danger" style={{ textAlign: 'center', paddingY: 96 }}>
			{results.errorMessage}
		</Label>
	{:else if item.kind === 'noResults'}
		<Label size="small" tone="dim" style={{ textAlign: 'center', paddingY: 96 }}>
			No results found for "{query}".
		</Label>
	{:else if item.kind === 'filters'}
		<View style={{ direction: 'row', wrap: true, gap: 8, paddingX: gutter.narrow }}>
			<FilterChip label="All" selected={results.filter === ''} onpress={() => results.select('')} />
			{#each results.filterTitles as title (title)}
				<FilterChip
					label={title}
					selected={results.filter === title}
					onpress={() => results.select(title)}
				/>
			{/each}
		</View>
	{:else if item.kind === 'filteredShelf'}
		<View style={{ paddingX: gutter.narrow }}>
			{#if shelfTracks(item.shelf)}
				{@render trackShelf(item.shelf, item.shelf.items.length, true)}
			{:else}
				{@render itemGrid(item.shelf.items)}
			{/if}
		</View>
	{:else if item.kind === 'topSection'}
		<View style={{ direction: 'row', gap: 24, paddingX: gutter.narrow }}>
			{#if results.topResult}
				<View style={{ flex: 2, minWidth: 0, gap: 8 }}>
					<Label size="heading" weight="bold">Top result</Label>
					<TopResultCard item={results.topResult} {context} />
				</View>
			{/if}
			{#if results.songsShelf}
				<View style={{ flex: 3, minWidth: 0, gap: 8 }}>
					<Label size="heading" weight="bold">Songs</Label>
					{@render trackShelf(results.songsShelf, topSongCount, false)}
				</View>
			{/if}
		</View>
	{:else if item.kind === 'shelf'}
		<MediaCardRow title={item.shelf.title} items={item.shelf.items} />
	{/if}
{/snippet}

<PageList items={blocks} key={(item) => item.key} rowGap={24} estimatedRowHeight={320}>
	{#snippet row(item)}
		{@render block(item)}
	{/snippet}
</PageList>
