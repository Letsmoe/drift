<!--
  One row of a long track list that is a block of a `PageList`: the row plays its track as part of the whole
  list, which becomes the queue.
-->
<script lang="ts">
	import type { ItemKind, Track } from '#lib/api/music.ts';
	import TrackTableRow from '#lib/components/TrackTableRow.svelte';
	import { player } from '#lib/player/player.svelte.ts';

	let {
		tracks,
		index,
		context,
		kind = 'song',
		showCover = true,
		showAlbum = true,
		dateAdded = undefined,
		onremove = undefined
	}: {
		tracks: Track[];
		index: number;
		/** Queue context, e.g. `playlist:<id>`. */
		context: string;
		kind?: ItemKind;
		showCover?: boolean;
		showAlbum?: boolean;
		dateAdded?: string;
		onremove?: (index: number) => void;
	} = $props();

	const removeHandler = $derived(removeHandlerFor());

	function removeHandlerFor(): (() => void) | undefined {
		if (!onremove) {
			return undefined;
		}
		return () => onremove(index);
	}
</script>

<TrackTableRow
	track={tracks[index]}
	{index}
	{kind}
	{showCover}
	{showAlbum}
	{dateAdded}
	onplay={() => player.playOrToggle(tracks[index], tracks, index, context, kind)}
	onremove={removeHandler}
/>
