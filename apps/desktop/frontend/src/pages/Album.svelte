<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { album } from '#lib/api/catalog.ts';
	import { createItem, createTrack, type AlbumPage, type Track } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import ArtistLinks from '#lib/components/ArtistLinks.svelte';
	import CollectionDownloadButton from '#lib/components/CollectionDownloadButton.svelte';
	import Label from '#lib/components/Label.svelte';
	import MoreMenuButton from '#lib/components/MoreMenuButton.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageList from '#lib/components/PageList.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import SaveButton from '#lib/components/SaveButton.svelte';
	import StickyPlayBar from '#lib/components/StickyPlayBar.svelte';
	import TrackTableHeader from '#lib/components/TrackTableHeader.svelte';
	import TrackTableRow from '#lib/components/TrackTableRow.svelte';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { formatSongCount, formatTotalDuration } from '#lib/format.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { collectionMenuEntries } from '#lib/menu/collectionMenuEntries.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { routeParameter } from '#lib/pages/routeParameter.ts';
	import type { PageProps } from '#lib/pages/pageProps.ts';
	import { playContext, shufflePlayContext } from '#lib/player/contextPlayback.ts';
	import { player } from '#lib/player/player.svelte.ts';

	type AlbumBlock =
		| { key: string; kind: 'header' }
		| { key: string; kind: 'actions' }
		| { key: string; kind: 'tableHeader' }
		| { key: string; kind: 'track'; index: number };

	// The play button has scrolled out of view once the header and the action row are above the window.
	const stickyBarScrollOffset = 280;

	let { params }: PageProps = $props();

	const resource = new RemoteResource(album);
	const albumId = $derived(routeParameter(params.id));
	const albumPage = $derived(resource.data);
	const context = $derived(`album:${albumId}`);
	const tracks = $derived(playableTracks(albumPage));
	const blocks = $derived(buildBlocks(tracks));
	let scrollOffset = $state(0);

	$effect(() => {
		resource.load(albumId);
		scrollOffset = 0;
	});

	$effect(() => {
		library.ensureRatings(tracks.map((track) => track.videoId));
		downloads.ensureStates(tracks.map((track) => track.videoId));
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

	function buildBlocks(listed: Track[]): AlbumBlock[] {
		const built: AlbumBlock[] = [
			{ key: 'header', kind: 'header' },
			{ key: 'actions', kind: 'actions' },
			{ key: 'tableHeader', kind: 'tableHeader' }
		];
		listed.forEach((_, index) => {
			built.push({ key: `track:${index}`, kind: 'track', index });
		});
		return built;
	}

	function metaLabel(loaded: AlbumPage): string {
		return `${loaded.year} • ${formatSongCount(loaded.tracks.length)}, ${formatTotalDuration(loaded.tracks)}`;
	}

	function playAt(index: number): void {
		player.playOrToggle(tracks[index], tracks, index, context, 'song');
	}
</script>

<View style={{ flex: 1, minHeight: 0 }}>
	<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

	{#if albumPage}
		<PageShell seed={albumPage.title} cover={albumPage.thumbnails}>
			<PageList
				items={blocks}
				key={(block) => block.key}
				paddingX={24}
				paddingTop={24}
				paddingBottom={24}
				estimatedRowHeight={49}
				onscroll={(position) => (scrollOffset = position.y)}
			>
				{#snippet row(block)}
					{#if block.kind === 'header'}
						<PageHeader label="Album" title={albumPage.title} thumbnails={albumPage.thumbnails}>
							{#snippet meta()}
								<ArtistLinks artists={albumPage.artists} color={appearance.muted} />
								<Label size="small">• {metaLabel(albumPage)}</Label>
							{/snippet}
						</PageHeader>
					{:else if block.kind === 'actions'}
						<PlaylistActionRow
							{context}
							title={albumPage.title}
							onplay={() => playContext(tracks, context)}
							onshuffle={() => shufflePlayContext(tracks, context)}
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
					{:else if block.kind === 'tableHeader'}
						<TrackTableHeader showAlbum={false} />
					{:else}
						<TrackTableRow
							track={tracks[block.index]}
							index={block.index}
							showCover={false}
							showAlbum={false}
							onplay={() => playAt(block.index)}
						/>
					{/if}
				{/snippet}
			</PageList>
			<StickyPlayBar
				visible={scrollOffset > stickyBarScrollOffset}
				{context}
				title={albumPage.title}
				onplay={() => playContext(tracks, context)}
			/>
		</PageShell>
	{/if}
</View>
