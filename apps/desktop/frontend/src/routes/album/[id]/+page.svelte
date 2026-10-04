<script lang="ts">
	import { page } from '$app/state';
	import { album } from '#lib/api/catalog.ts';
	import { createItem, createTrack, type AlbumPage, type Track } from '#lib/api/music.ts';
	import CollectionDownloadButton from '#lib/components/CollectionDownloadButton.svelte';
	import ArtistLinks from '#lib/components/ArtistLinks.svelte';
	import MoreMenuButton from '#lib/components/MoreMenuButton.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import SaveButton from '#lib/components/SaveButton.svelte';
	import TrackTable from '#lib/components/TrackTable.svelte';
	import { formatSongCount, formatTotalDuration } from '#lib/format.ts';
	import { collectionMenuEntries } from '#lib/menu/collectionMenuEntries.ts';
	import { playContext } from '#lib/player/contextPlayback.ts';
	import { routeParameter } from '#lib/pages/routeParameter.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';

	const resource = new RemoteResource(album);
	const albumId = $derived(routeParameter(page.params.id));
	const albumPage = $derived(resource.data);
	const context = $derived(`album:${albumId}`);
	const tracks = $derived(playableTracks(albumPage));

	$effect(() => {
		resource.load(albumId);
	});

	// Album rows carry no cover or album of their own; the queue needs both.
	function playableTracks(loaded: AlbumPage | undefined): Track[] {
		if (!loaded) {
			return [];
		}
		return loaded.tracks.map((track) => withAlbumInfo(track, loaded));
	}

	function withAlbumInfo(track: Track, loaded: AlbumPage): Track {
		return createTrack(track, {
			thumbnails: loaded.thumbnails,
			album: { id: albumId, name: loaded.title }
		});
	}

	function metaLabel(loaded: AlbumPage): string {
		return `${loaded.year} • ${formatSongCount(loaded.tracks.length)}, ${formatTotalDuration(loaded.tracks)}`;
	}
</script>

<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

{#if albumPage}
	<PageShell seed={albumPage.title}>
		<PageHeader label="Album" title={albumPage.title} thumbnails={albumPage.thumbnails}>
			{#snippet meta()}
				<ArtistLinks artists={albumPage.artists} />
				<span>• {metaLabel(albumPage)}</span>
			{/snippet}
		</PageHeader>
		<PlaylistActionRow
			{context}
			title={albumPage.title}
			onplay={() => playContext(tracks, context)}
		>
			<SaveButton
				item={createItem(
					'album',
					albumId,
					albumPage.title,
					albumPage.subtitle,
					albumPage.thumbnails
				)}
			/>
			<CollectionDownloadButton kind="album" id={albumId} title={albumPage.title} />
			<MoreMenuButton buildEntries={() => collectionMenuEntries(tracks)} />
		</PlaylistActionRow>
		<TrackTable {tracks} {context} showCover={false} showAlbum={false} />
	</PageShell>
{/if}
