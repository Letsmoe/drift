<script lang="ts">
	import { page } from '$app/state';
	import { playlist, playlistContinuation } from '#lib/api/catalog.ts';
	import { createItem, type PlaylistPage, type Track } from '#lib/api/music.ts';
	import CollectionDownloadButton from '#lib/components/CollectionDownloadButton.svelte';
	import LoadMoreSentinel from '#lib/components/LoadMoreSentinel.svelte';
	import MoreMenuButton from '#lib/components/MoreMenuButton.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import SaveButton from '#lib/components/SaveButton.svelte';
	import TrackListControls from '#lib/components/TrackListControls.svelte';
	import TrackTable from '#lib/components/TrackTable.svelte';
	import { formatSongCount, formatTotalDuration } from '#lib/format.ts';
	import { collectionMenuEntries } from '#lib/menu/collectionMenuEntries.ts';
	import { ContinuationList } from '#lib/pages/continuationList.svelte.ts';
	import { TrackListFilter } from '#lib/pages/trackListFilter.svelte.ts';
	import { routeParameter } from '#lib/pages/routeParameter.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { playContext, shufflePlayContext } from '#lib/player/contextPlayback.ts';

	const resource = new RemoteResource(playlist);
	const trackList = new ContinuationList<Track>(fetchMoreTracks);
	const playlistId = $derived(routeParameter(page.params.id));
	const playlistPage = $derived(resource.data);
	const context = $derived(`playlist:${playlistId}`);
	const filter = new TrackListFilter();
	const shownTracks = $derived(filter.apply(trackList.items));

	$effect(() => {
		resource.load(playlistId);
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

	async function fetchMoreTracks(token: string) {
		const next = await playlistContinuation(token);
		return { items: next.tracks, continuation: next.continuation };
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

<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

{#if playlistPage}
	<PageShell seed={playlistPage.title}>
		<PageHeader
			label="Playlist"
			title={playlistPage.title}
			thumbnails={playlistPage.thumbnails}
			description={playlistPage.description}
		>
			{#snippet meta()}
				<span>{metaLabel(playlistPage)}</span>
			{/snippet}
		</PageHeader>
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
			<CollectionDownloadButton kind="playlist" id={playlistId} title={playlistPage.title} />
			<MoreMenuButton buildEntries={() => collectionMenuEntries(trackList.items)} />
		</PlaylistActionRow>
		<TrackListControls {filter} />
		<TrackTable tracks={shownTracks} {context} />
		{#if filter.active && shownTracks.length === 0 && !trackList.loadingMore}
			<p class="py-12 text-center text-sm text-dim">No songs match “{filter.query}”.</p>
		{/if}
		{#if trackList.hasMore && !filter.active}
			<LoadMoreSentinel itemCount={trackList.items.length} onvisible={() => trackList.loadMore()} />
		{/if}
	</PageShell>
{/if}
