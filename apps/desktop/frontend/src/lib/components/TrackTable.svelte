<!--
  A short list of tracks with its column header, all built at once. Long lists (albums, playlists) are rows of
  a `PageList` instead, built from `TrackTableRow` directly.
-->
<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import type { ItemKind, Track } from '#lib/api/music.ts';
	import TrackTableHeader from '#lib/components/TrackTableHeader.svelte';
	import TrackTableRow from '#lib/components/TrackTableRow.svelte';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';

	let {
		tracks,
		context,
		kind = 'song',
		showHeader = true,
		showCover = true,
		showAlbum = true,
		visibleCount = tracks.length,
		dateAddedLabels = undefined,
		onremove = undefined
	}: {
		tracks: Track[];
		/** Queue context, e.g. `album:<id>`; the whole list becomes the queue. */
		context: string;
		kind?: ItemKind;
		showHeader?: boolean;
		showCover?: boolean;
		showAlbum?: boolean;
		/** Shows only the first rows while the full list stays the queue. */
		visibleCount?: number;
		/** One label per track; shows the "Date added" column. */
		dateAddedLabels?: string[];
		onremove?: (index: number) => void;
	} = $props();

	const visibleTracks = $derived(tracks.slice(0, visibleCount));
	const showDateAdded = $derived(dateAddedLabels !== undefined);
	// Rows only show linked albums; a column where none is linked would stay empty.
	const hasAlbumColumn = $derived(
		showAlbum && tracks.some((track) => track.album !== undefined && track.album.id !== '')
	);

	$effect(() => {
		library.ensureRatings(visibleTracks.map((track) => track.videoId));
		downloads.ensureStates(visibleTracks.map((track) => track.videoId));
	});

	function playAt(index: number): void {
		player.playOrToggle(tracks[index], tracks, index, context, kind);
	}

	function dateAddedAt(index: number): string | undefined {
		return dateAddedLabels?.[index];
	}

	function removeHandlerFor(index: number): (() => void) | undefined {
		if (!onremove) {
			return undefined;
		}
		return () => onremove(index);
	}
</script>

<View>
	{#if showHeader}
		<TrackTableHeader showAlbum={hasAlbumColumn} {showDateAdded} />
	{/if}
	{#each visibleTracks as track, index (index)}
		<TrackTableRow
			{track}
			{index}
			{kind}
			{showCover}
			showAlbum={hasAlbumColumn}
			dateAdded={dateAddedAt(index)}
			onplay={() => playAt(index)}
			onremove={removeHandlerFor(index)}
		/>
	{/each}
</View>
