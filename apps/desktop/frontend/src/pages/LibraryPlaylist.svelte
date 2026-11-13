<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { libraryPlaylist, type Playlist } from '#lib/api/library.ts';
	import type { Thumbnail, Track } from '#lib/api/music.ts';
	import CollectionDownloadButton from '#lib/components/CollectionDownloadButton.svelte';
	import ConfirmDialog from '#lib/components/ConfirmDialog.svelte';
	import EditPlaylistDialog from '#lib/components/EditPlaylistDialog.svelte';
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
	import {
		formatDate,
		formatSongCount,
		formatTotalDuration,
		hasLinkedAlbums
	} from '#lib/format.ts';
	import { pencilSimple, trash } from '#lib/icons.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { collectionMenuEntries } from '#lib/menu/collectionMenuEntries.ts';
	import type { MenuEntry } from '#lib/menu/menuEntry.ts';
	import type { PageProps } from '#lib/pages/pageProps.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { routeParameter } from '#lib/pages/routeParameter.ts';
	import { ensureTrackStates } from '#lib/pages/trackStates.ts';
	import { playContext, shufflePlayContext } from '#lib/player/contextPlayback.ts';
	import { goto } from '#lib/router.svelte.ts';

	type OpenDialog = 'none' | 'edit' | 'delete';
	type PlaylistBlock =
		| { key: string; kind: 'header' }
		| { key: string; kind: 'actions' }
		| { key: string; kind: 'empty' }
		| { key: string; kind: 'tableHeader' }
		| { key: string; kind: 'track'; index: number };

	// The play button has scrolled out of view once the header and the action row are above the window.
	const stickyBarScrollOffset = 280;

	let { params }: PageProps = $props();

	const resource = new RemoteResource(libraryPlaylist);
	const playlistId = $derived(routeParameter(params.id));
	const context = $derived(`library:${playlistId}`);
	const playlist = $derived(resource.data);
	const tracks = $derived(trackList(playlist));
	const dateAddedLabels = $derived(addedDates(playlist));
	const showAlbum = $derived(hasLinkedAlbums(tracks));
	const blocks = $derived(buildBlocks());
	let openDialog = $state<OpenDialog>('none');
	let scrollOffset = $state(0);

	$effect(() => {
		resource.load(playlistId);
		scrollOffset = 0;
	});

	$effect(() => {
		ensureTrackStates(tracks);
	});

	function trackList(loaded: Playlist | undefined): Track[] {
		if (!loaded) {
			return [];
		}
		return loaded.entries.map((entry) => entry.track);
	}

	function addedDates(loaded: Playlist | undefined): string[] {
		if (!loaded) {
			return [];
		}
		return loaded.entries.map((entry) => formatDate(entry.addedAt));
	}

	function buildBlocks(): PlaylistBlock[] {
		const built: PlaylistBlock[] = [
			{ key: 'header', kind: 'header' },
			{ key: 'actions', kind: 'actions' }
		];
		if (tracks.length === 0) {
			built.push({ key: 'empty', kind: 'empty' });
			return built;
		}
		built.push({ key: 'tableHeader', kind: 'tableHeader' });
		tracks.forEach((_, index) => {
			built.push({ key: `track:${playlist?.entries[index].id}`, kind: 'track', index });
		});
		return built;
	}

	function coverThumbnails(): Thumbnail[] {
		if (tracks.length === 0) {
			return [];
		}
		return tracks[0].thumbnails;
	}

	function metaLabel(): string {
		if (tracks.length === 0) {
			return formatSongCount(0);
		}
		return `${formatSongCount(tracks.length)}, ${formatTotalDuration(tracks)}`;
	}

	function menuEntries(): MenuEntry[] {
		return [
			{
				label: 'Edit details',
				icon: pencilSimple,
				onselect: () => (openDialog = 'edit')
			},
			...collectionMenuEntries(tracks),
			{
				label: 'Delete playlist',
				icon: trash,
				separatorBefore: true,
				onselect: () => (openDialog = 'delete')
			}
		];
	}

	async function removeEntry(index: number): Promise<void> {
		if (!playlist) {
			return;
		}
		await library.removeFromPlaylist(playlist.id, playlist.entries[index].id);
		await resource.reload();
	}

	async function saveDetails(name: string, description: string): Promise<void> {
		await library.renamePlaylist(playlistId, name, description);
		await resource.reload();
	}

	async function deletePlaylist(): Promise<void> {
		await library.deletePlaylist(playlistId);
		goto('/');
	}
</script>

<View style={{ flex: 1, minHeight: 0 }}>
	<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

	{#if playlist}
		<PageShell seed={playlist.name}>
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
						<PageHeader label="Playlist" title={playlist.name} thumbnails={coverThumbnails()}>
							{#snippet meta()}
								<Label size="small">{metaLabel()}</Label>
							{/snippet}
						</PageHeader>
					{:else if block.kind === 'actions'}
						<PlaylistActionRow
							{context}
							title={playlist.name}
							onplay={() => playContext(tracks, context)}
							onshuffle={() => shufflePlayContext(tracks, context)}
						>
							<CollectionDownloadButton kind="local" id={playlistId} title={playlist.name} />
							<MoreMenuButton buildEntries={menuEntries} />
						</PlaylistActionRow>
					{:else if block.kind === 'empty'}
						<EmptyMessage>
							This playlist is empty. Add songs from the "..." menu of any track.
						</EmptyMessage>
					{:else if block.kind === 'tableHeader'}
						<TrackTableHeader {showAlbum} showDateAdded />
					{:else}
						<TrackListRow
							{tracks}
							index={block.index}
							{context}
							{showAlbum}
							dateAdded={dateAddedLabels[block.index]}
							onremove={removeEntry}
						/>
					{/if}
				{/snippet}
			</PageList>
			<StickyPlayBar
				visible={scrollOffset > stickyBarScrollOffset}
				{context}
				title={playlist.name}
				onplay={() => playContext(tracks, context)}
			/>
		</PageShell>

		{#if openDialog === 'edit'}
			<EditPlaylistDialog
				name={playlist.name}
				description={playlist.description}
				onsave={saveDetails}
				onclose={() => (openDialog = 'none')}
			/>
		{/if}
		{#if openDialog === 'delete'}
			<ConfirmDialog
				title="Delete playlist"
				message={`Delete "${playlist.name}" from your library? This cannot be undone.`}
				confirmLabel="Delete"
				onconfirm={deletePlaylist}
				onclose={() => (openDialog = 'none')}
			/>
		{/if}
	{/if}
</View>
