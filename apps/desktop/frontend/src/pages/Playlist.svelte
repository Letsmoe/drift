<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { playlist, playlistContinuation } from '#lib/api/catalog.ts';
	import { createItem, type PlaylistPage, type Track } from '#lib/api/music.ts';
	import CollectionDownloadButton from '#lib/components/CollectionDownloadButton.svelte';
	import EmptyMessage from '#lib/components/EmptyMessage.svelte';
	import Label from '#lib/components/Label.svelte';
	import LoadMoreSentinel from '#lib/components/LoadMoreSentinel.svelte';
	import MoreMenuButton from '#lib/components/MoreMenuButton.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageList from '#lib/components/PageList.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import SaveButton from '#lib/components/SaveButton.svelte';
	import StickyPlayBar from '#lib/components/StickyPlayBar.svelte';
	import TrackListControls from '#lib/components/TrackListControls.svelte';
	import TrackListRow from '#lib/components/TrackListRow.svelte';
	import TrackTableHeader from '#lib/components/TrackTableHeader.svelte';
	import { trackRowHeight } from '#lib/components/trackColumns.ts';
	import { formatSongCount, formatTotalDuration, hasLinkedAlbums } from '#lib/format.ts';
	import { collectionMenuEntries } from '#lib/menu/collectionMenuEntries.ts';
	import { ContinuationList } from '#lib/pages/continuationList.svelte.ts';
	import type { PageProps } from '#lib/pages/pageProps.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { routeParameter } from '#lib/pages/routeParameter.ts';
	import { TrackListFilter } from '#lib/pages/trackListFilter.svelte.ts';
	import { ensureTrackStates } from '#lib/pages/trackStates.ts';
	import { playContext, shufflePlayContext } from '#lib/player/contextPlayback.ts';

	type PlaylistBlock =
		| { key: string; kind: 'header' }
		| { key: string; kind: 'actions' }
		| { key: string; kind: 'controls' }
		| { key: string; kind: 'tableHeader' }
		| { key: string; kind: 'track'; index: number }
		| { key: string; kind: 'noMatch' }
		| { key: string; kind: 'loadMore' };

	// The play button has scrolled out of view once the header and the action row are above the window.
	const stickyBarScrollOffset = 280;

	let { params }: PageProps = $props();

	const resource = new RemoteResource(playlist);
	const trackList = new ContinuationList<Track>(fetchMoreTracks);
	const filter = new TrackListFilter();
	const playlistId = $derived(routeParameter(params.id));
	const playlistPage = $derived(resource.data);
	const context = $derived(`playlist:${playlistId}`);
	const shownTracks = $derived(filter.apply(trackList.items));
	const showAlbum = $derived(hasLinkedAlbums(trackList.items));
	const blocks = $derived(buildBlocks());
	let scrollOffset = $state(0);

	$effect(() => {
		resource.load(playlistId);
		scrollOffset = 0;
	});

	$effect(() => {
		if (playlistPage) {
			trackList.reset(playlistPage.tracks, playlistPage.continuation);
		}
	});

	// Search and sort cover the whole playlist, not only the pages scrolled so far.
	$effect(() => {
		if (filter.active) {
			trackList.loadAll();
		}
	});

	$effect(() => {
		ensureTrackStates(shownTracks);
	});

	async function fetchMoreTracks(token: string) {
		const next = await playlistContinuation(token);
		return { items: next.tracks, continuation: next.continuation };
	}

	function buildBlocks(): PlaylistBlock[] {
		const built: PlaylistBlock[] = [
			{ key: 'header', kind: 'header' },
			{ key: 'actions', kind: 'actions' },
			{ key: 'controls', kind: 'controls' },
			{ key: 'tableHeader', kind: 'tableHeader' }
		];
		shownTracks.forEach((_, index) => {
			built.push({ key: `track:${index}`, kind: 'track', index });
		});
		if (filter.active && shownTracks.length === 0 && !trackList.loadingMore) {
			built.push({ key: 'noMatch', kind: 'noMatch' });
		}
		if (trackList.hasMore && !filter.active) {
			built.push({ key: `loadMore:${trackList.items.length}`, kind: 'loadMore' });
		}
		return built;
	}

	function metaLabel(loaded: PlaylistPage): string {
		const songs = formatSongCount(trackList.items.length);
		const length = formatTotalDuration(trackList.items);
		if (trackList.hasMore) {
			return `${loaded.author} • ${songs} so far`;
		}
		return `${loaded.author} • ${songs}, ${length}`;
	}
</script>

<View style={{ flex: 1, minHeight: 0 }}>
	<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

	{#if playlistPage}
		<PageShell seed={playlistPage.title}>
			<PageList
				items={blocks}
				key={(block) => block.key}
				paddingX={24}
				paddingTop={24}
				paddingBottom={24}
				estimatedRowHeight={trackRowHeight}
				onscroll={(position) => (scrollOffset = position.y)}
			>
				{#snippet row(block)}
					{#if block.kind === 'header'}
						<PageHeader
							label="Playlist"
							title={playlistPage.title}
							thumbnails={playlistPage.thumbnails}
							description={playlistPage.description}
						>
							{#snippet meta()}
								<Label size="small">{metaLabel(playlistPage)}</Label>
							{/snippet}
						</PageHeader>
					{:else if block.kind === 'actions'}
						<PlaylistActionRow
							{context}
							title={playlistPage.title}
							onplay={() => playContext(shownTracks, context)}
							onshuffle={() => shufflePlayContext(shownTracks, context)}
						>
							<SaveButton
								item={createItem(
									'playlist',
									playlistId,
									playlistPage.title,
									playlistPage.author,
									playlistPage.thumbnails
								)}
							/>
							<CollectionDownloadButton
								kind="playlist"
								id={playlistId}
								title={playlistPage.title}
							/>
							<MoreMenuButton buildEntries={() => collectionMenuEntries(trackList.items)} />
						</PlaylistActionRow>
					{:else if block.kind === 'controls'}
						<TrackListControls {filter} />
					{:else if block.kind === 'tableHeader'}
						<TrackTableHeader {showAlbum} />
					{:else if block.kind === 'track'}
						<TrackListRow tracks={shownTracks} index={block.index} {context} {showAlbum} />
					{:else if block.kind === 'noMatch'}
						<EmptyMessage>No songs match “{filter.query}”.</EmptyMessage>
					{:else}
						<LoadMoreSentinel onvisible={() => trackList.loadMore()} />
					{/if}
				{/snippet}
			</PageList>
			<StickyPlayBar
				visible={scrollOffset > stickyBarScrollOffset}
				{context}
				title={playlistPage.title}
				onplay={() => playContext(shownTracks, context)}
			/>
		</PageShell>
	{/if}
</View>
