<script lang="ts">
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import type { Track } from '#lib/api/music.ts';
	import { likedCollectionId } from '#lib/api/downloads.ts';
	import { likedTracks } from '#lib/api/library.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import CollectionDownloadButton from '#lib/components/CollectionDownloadButton.svelte';
	import MoreMenuButton from '#lib/components/MoreMenuButton.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import TrackTable from '#lib/components/TrackTable.svelte';
	import { formatSongCount } from '#lib/format.ts';
	import { collectionMenuEntries } from '#lib/menu/collectionMenuEntries.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { playContext } from '#lib/player/contextPlayback.ts';

	const context = 'liked';
	const resource = new RemoteResource(() => likedTracks());
	const tracks = $derived(stillLikedTracks(resource.data));

	// A track whose rating changed away from like drops out without refetching.
	function stillLikedTracks(data: Track[] | undefined): Track[] {
		if (!data) {
			return [];
		}
		return data.filter(isStillLiked);
	}

	function isStillLiked(track: Track): boolean {
		const rating = library.ratings.get(track.videoId);
		return rating === undefined || rating === 1;
	}

	$effect(() => {
		resource.load('liked');
	});
</script>

<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

{#if resource.data}
	<PageShell seed="Liked Songs">
		<PageHeader label="Playlist" title="Liked Songs">
			{#snippet cover()}
				<div class="flex size-full items-center justify-center bg-hover text-default">
					<HeartIcon size={96} weight="fill" />
				</div>
			{/snippet}
			{#snippet meta()}
				<span>{formatSongCount(tracks.length)}</span>
			{/snippet}
		</PageHeader>
		<PlaylistActionRow {context} title="Liked Songs" onplay={() => playContext(tracks, context)}>
			<CollectionDownloadButton kind="likes" id={likedCollectionId} title="Liked Songs" />
			<MoreMenuButton buildEntries={() => collectionMenuEntries(tracks)} />
		</PlaylistActionRow>
		{#if tracks.length === 0}
			<p class="py-12 text-center text-sm text-dim">Songs you like will appear here.</p>
		{:else}
			<TrackTable {tracks} {context} />
		{/if}
	</PageShell>
{/if}
