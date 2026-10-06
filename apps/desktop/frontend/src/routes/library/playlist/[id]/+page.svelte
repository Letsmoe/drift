<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import PencilSimpleIcon from 'phosphor-svelte/lib/PencilSimpleIcon';
	import TrashIcon from 'phosphor-svelte/lib/TrashIcon';
	import { libraryPlaylist, type Playlist } from '#lib/api/library.ts';
	import type { Thumbnail, Track } from '#lib/api/music.ts';
	import CollectionDownloadButton from '#lib/components/CollectionDownloadButton.svelte';
	import ConfirmDialog from '#lib/components/ConfirmDialog.svelte';
	import EditPlaylistDialog from '#lib/components/EditPlaylistDialog.svelte';
	import MoreMenuButton from '#lib/components/MoreMenuButton.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import TrackTable from '#lib/components/TrackTable.svelte';
	import { formatDate, formatSongCount, formatTotalDuration } from '#lib/format.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { collectionMenuEntries } from '#lib/menu/collectionMenuEntries.ts';
	import type { MenuEntry } from '#lib/menu/menuEntry.ts';
	import { routeParameter } from '#lib/pages/routeParameter.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { playContext, shufflePlayContext } from '#lib/player/contextPlayback.ts';

	type OpenDialog = 'none' | 'edit' | 'delete';

	const resource = new RemoteResource(libraryPlaylist);
	const playlistId = $derived(routeParameter(page.params.id));
	const context = $derived(`library:${playlistId}`);
	const playlist = $derived(resource.data);
	const tracks = $derived(trackList(playlist));
	const dateAddedLabels = $derived(addedDates(playlist));
	let openDialog = $state<OpenDialog>('none');

	$effect(() => {
		resource.load(playlistId);
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
				icon: PencilSimpleIcon,
				onselect: () => (openDialog = 'edit')
			},
			...collectionMenuEntries(tracks),
			{
				label: 'Delete playlist',
				icon: TrashIcon,
				separatorBefore: true,
				onselect: () => (openDialog = 'delete')
			}
		];
	}

	async function removeEntry(index: number) {
		if (!playlist) {
			return;
		}
		await library.removeFromPlaylist(playlist.id, playlist.entries[index].id);
		await resource.reload();
	}

	async function saveDetails(name: string, description: string) {
		await library.renamePlaylist(playlistId, name, description);
		await resource.reload();
	}

	async function deletePlaylist() {
		await library.deletePlaylist(playlistId);
		await goto('/');
	}
</script>

<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

{#if playlist}
	<PageShell seed={playlist.name}>
		<PageHeader label="Playlist" title={playlist.name} thumbnails={coverThumbnails()}>
			{#snippet meta()}
				<span>{metaLabel()}</span>
			{/snippet}
		</PageHeader>
		<PlaylistActionRow
			{context}
			title={playlist.name}
			onplay={() => playContext(tracks, context)}
			onshuffle={() => shufflePlayContext(tracks, context)}
		>
			<CollectionDownloadButton kind="local" id={playlistId} title={playlist.name} />
			<MoreMenuButton buildEntries={menuEntries} />
		</PlaylistActionRow>
		{#if tracks.length === 0}
			<p class="py-12 text-center text-sm text-dim">
				This playlist is empty. Add songs from the "..." menu of any track.
			</p>
		{:else}
			<TrackTable {tracks} {context} {dateAddedLabels} onremove={removeEntry} />
		{/if}
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
			message="Delete &quot;{playlist.name}&quot; from your library? This cannot be undone."
			confirmLabel="Delete"
			onconfirm={deletePlaylist}
			onclose={() => (openDialog = 'none')}
		/>
	{/if}
{/if}
