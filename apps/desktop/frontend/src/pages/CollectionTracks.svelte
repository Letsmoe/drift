<script lang="ts">
	import { Icon, View } from '@neoworks/svelte-gpui/ui';
	import { likedCollectionId } from '#lib/api/downloads.ts';
	import { likedTracks } from '#lib/api/library.ts';
	import type { Track } from '#lib/api/music.ts';
	import CollectionDownloadButton from '#lib/components/CollectionDownloadButton.svelte';
	import EmptyMessage from '#lib/components/EmptyMessage.svelte';
	import Label from '#lib/components/Label.svelte';
	import MoreMenuButton from '#lib/components/MoreMenuButton.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageList from '#lib/components/PageList.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import StickyPlayBar from '#lib/components/StickyPlayBar.svelte';
	import TrackListRow from '#lib/components/TrackListRow.svelte';
	import TrackTableHeader from '#lib/components/TrackTableHeader.svelte';
	import { trackRowHeight } from '#lib/components/trackColumns.ts';
	import { formatSongCount, hasLinkedAlbums } from '#lib/format.ts';
	import { heartFill } from '#lib/icons.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { collectionMenuEntries } from '#lib/menu/collectionMenuEntries.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { ensureTrackStates } from '#lib/pages/trackStates.ts';
	import { playContext, shufflePlayContext } from '#lib/player/contextPlayback.ts';
	import { colors } from '#lib/theme.ts';

	type LikedBlock =
		| { key: string; kind: 'header' }
		| { key: string; kind: 'actions' }
		| { key: string; kind: 'empty' }
		| { key: string; kind: 'tableHeader' }
		| { key: string; kind: 'track'; index: number };

	const context = 'liked';
	// The play button has scrolled out of view once the header and the action row are above the window.
	const stickyBarScrollOffset = 280;

	const resource = new RemoteResource(() => likedTracks());
	const tracks = $derived(stillLikedTracks(resource.data));
	const showAlbum = $derived(hasLinkedAlbums(tracks));
	const blocks = $derived(buildBlocks());
	let scrollOffset = $state(0);

	$effect(() => {
		resource.load('liked');
	});

	$effect(() => {
		ensureTrackStates(tracks);
	});

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

	function buildBlocks(): LikedBlock[] {
		const built: LikedBlock[] = [
			{ key: 'header', kind: 'header' },
			{ key: 'actions', kind: 'actions' }
		];
		if (tracks.length === 0) {
			built.push({ key: 'empty', kind: 'empty' });
			return built;
		}
		built.push({ key: 'tableHeader', kind: 'tableHeader' });
		tracks.forEach((track, index) => {
			built.push({ key: `track:${track.videoId}`, kind: 'track', index });
		});
		return built;
	}
</script>

<View style={{ flex: 1, minHeight: 0 }}>
	<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

	{#if resource.data}
		<PageShell seed="Liked Songs">
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
						<PageHeader label="Playlist" title="Liked Songs">
							{#snippet cover()}
								<View
									style={{
										size: 'full',
										align: 'center',
										justify: 'center',
										background: colors.hover,
										color: colors.text
									}}
								>
									<Icon svg={heartFill} size={96} />
								</View>
							{/snippet}
							{#snippet meta()}
								<Label size="small">{formatSongCount(tracks.length)}</Label>
							{/snippet}
						</PageHeader>
					{:else if block.kind === 'actions'}
						<PlaylistActionRow
							{context}
							title="Liked Songs"
							onplay={() => playContext(tracks, context)}
							onshuffle={() => shufflePlayContext(tracks, context)}
						>
							<CollectionDownloadButton kind="likes" id={likedCollectionId} title="Liked Songs" />
							<MoreMenuButton buildEntries={() => collectionMenuEntries(tracks)} />
						</PlaylistActionRow>
					{:else if block.kind === 'empty'}
						<EmptyMessage>Songs you like will appear here.</EmptyMessage>
					{:else if block.kind === 'tableHeader'}
						<TrackTableHeader {showAlbum} />
					{:else}
						<TrackListRow {tracks} index={block.index} {context} {showAlbum} />
					{/if}
				{/snippet}
			</PageList>
			<StickyPlayBar
				visible={scrollOffset > stickyBarScrollOffset}
				{context}
				title="Liked Songs"
				onplay={() => playContext(tracks, context)}
			/>
		</PageShell>
	{/if}
</View>
