<script lang="ts">
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import type { Track } from '#lib/api/music.ts';
	import { likedTracks } from '#lib/api/library.ts';
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
	const tracks = $derived(loadedTracks(resource.data));

	function loadedTracks(data: Track[] | undefined): Track[] {
		if (!data) {
			return [];
		}
		return data;
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
				<div
					class="flex size-full items-center justify-center bg-linear-to-br from-violet to-blue text-white"
				>
					<HeartIcon size={96} weight="fill" />
				</div>
			{/snippet}
			{#snippet meta()}
				<span>{formatSongCount(tracks.length)}</span>
			{/snippet}
		</PageHeader>
		<PlaylistActionRow {context} title="Liked Songs" onplay={() => playContext(tracks, context)}>
			<MoreMenuButton buildEntries={() => collectionMenuEntries(tracks)} />
		</PlaylistActionRow>
		{#if tracks.length === 0}
			<p class="py-12 text-center text-sm text-dim">Songs you like will appear here.</p>
		{:else}
			<TrackTable {tracks} {context} />
		{/if}
	</PageShell>
{/if}
