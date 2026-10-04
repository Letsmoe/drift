<script lang="ts">
	import ClockIcon from 'phosphor-svelte/lib/ClockIcon';
	import type { ItemKind, Track } from '#lib/api/music.ts';
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
	const hasAlbumColumn = $derived(showAlbum && tracks.some((track) => track.album !== undefined));
	const gridClass = $derived(gridClassFor(hasAlbumColumn, showDateAdded));

	$effect(() => {
		library.ensureRatings(visibleTracks.map((track) => track.videoId));
		downloads.ensureStates(visibleTracks.map((track) => track.videoId));
	});

	function gridClassFor(withAlbum: boolean, withDate: boolean): string {
		if (withAlbum && withDate) {
			return 'grid-cols-[1rem_minmax(0,6fr)_minmax(0,4fr)_minmax(0,3fr)_minmax(8rem,1fr)]';
		}
		if (withAlbum) {
			return 'grid-cols-[1rem_minmax(0,6fr)_minmax(0,4fr)_minmax(8rem,1fr)]';
		}
		if (withDate) {
			return 'grid-cols-[1rem_minmax(0,6fr)_minmax(0,3fr)_minmax(8rem,1fr)]';
		}
		return 'grid-cols-[1rem_minmax(0,1fr)_minmax(8rem,auto)]';
	}

	function playAt(index: number) {
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

<div role="table" class="flex flex-col">
	{#if showHeader}
		<div
			role="row"
			class={[
				'sticky top-0 z-10 mb-2 grid h-9 items-center gap-4 border-b border-line bg-surface px-4 text-sm text-dim',
				gridClass
			]}
		>
			<span class="text-right">#</span>
			<span>Title</span>
			{#if hasAlbumColumn}<span>Album</span>{/if}
			{#if showDateAdded}<span>Date added</span>{/if}
			<span class="flex justify-end pr-9"><ClockIcon size={16} /></span>
		</div>
	{/if}
	{#each visibleTracks as track, index (index)}
		<TrackTableRow
			{track}
			{index}
			{gridClass}
			{kind}
			{showCover}
			showAlbum={hasAlbumColumn}
			dateAdded={dateAddedAt(index)}
			onplay={() => playAt(index)}
			onremove={removeHandlerFor(index)}
		/>
	{/each}
</div>
