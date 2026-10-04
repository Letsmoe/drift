<script lang="ts">
	import { page } from '$app/state';
	import { playlist, playlistContinuation } from '#lib/api/catalog.ts';
	import { createItem, type PlaylistPage, type Track } from '#lib/api/music.ts';
	import LoadMoreSentinel from '#lib/components/LoadMoreSentinel.svelte';
	import MoreMenuButton from '#lib/components/MoreMenuButton.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import SaveButton from '#lib/components/SaveButton.svelte';
	import TrackTable from '#lib/components/TrackTable.svelte';
	import { formatSongCount } from '#lib/format.ts';
	import { collectionMenuEntries } from '#lib/menu/collectionMenuEntries.ts';
	import { ContinuationList } from '#lib/pages/continuationList.svelte.ts';
	import { routeParameter } from '#lib/pages/routeParameter.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { playContext } from '#lib/player/contextPlayback.ts';

	const resource = new RemoteResource(playlist);
	const trackList = new ContinuationList<Track>(fetchMoreTracks);
	const playlistId = $derived(routeParameter(page.params.id));
	const playlistPage = $derived(resource.data);
	const context = $derived(`playlist:${playlistId}`);

	$effect(() => {
		resource.load(playlistId);
	});

	$effect(() => {
		if (playlistPage) {
			trackList.reset(playlistPage.tracks, playlistPage.continuation);
		}
	});

	async function fetchMoreTracks(token: string) {
		const next = await playlistContinuation(token);
		return { items: next.tracks, continuation: next.continuation };
	}

	function metaLabel(loaded: PlaylistPage): string {
		return `${loaded.author} • ${formatSongCount(trackList.items.length)}`;
	}
</script>

<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

{#if playlistPage}
	<PageShell seed={playlistPage.title}>
		<PageHeader label="Playlist" title={playlistPage.title} thumbnails={playlistPage.thumbnails}>
			{#snippet meta()}
				<span>{metaLabel(playlistPage)}</span>
			{/snippet}
		</PageHeader>
		<PlaylistActionRow
			{context}
			title={playlistPage.title}
			onplay={() => playContext(trackList.items, context)}
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
			<MoreMenuButton buildEntries={() => collectionMenuEntries(trackList.items)} />
		</PlaylistActionRow>
		<TrackTable tracks={trackList.items} {context} />
		{#if trackList.hasMore}
			<LoadMoreSentinel itemCount={trackList.items.length} onvisible={() => trackList.loadMore()} />
		{/if}
	</PageShell>
{/if}
