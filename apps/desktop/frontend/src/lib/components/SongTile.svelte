<script lang="ts">
	import type { Track } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import TrackContextMenu from '#lib/components/TrackContextMenu.svelte';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';
	import { formatTrackByline, trackHref, trackKindOf } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { whenVisible } from '#lib/whenVisible.ts';

	let {
		track,
		tracks = [track],
		index = 0,
		context = 'track',
		onplay,
		onvisible
	}: {
		track: Track;
		tracks?: Track[];
		index?: number;
		context?: string;
		onplay?: () => void;
		onvisible?: () => void;
	} = $props();

	const contextMenu = new PopupMenu();
	const current = $derived(player.isCurrent(track));
	const playing = $derived(player.isPlaying(track));
	const pageHref = $derived(trackHref(track));

	function play(): void {
		onplay?.();
		player.playOrToggle(track, tracks, index, context, trackKindOf(track));
	}

	function reportVisible(): void {
		onvisible?.();
	}
</script>

<div
	role="group"
	aria-label={track.title}
	class={[
		'group/tile duration-fast relative flex h-14 items-center gap-3 overflow-hidden rounded-md pr-2 transition-colors',
		current && 'bg-hover ring-1 ring-action/40',
		!current && 'bg-raised hover:bg-hover'
	]}
	{@attach whenVisible(reportVisible)}
	oncontextmenu={(event) => contextMenu.openAtPointer(event)}
>
	<Cover thumbnails={track.thumbnails} size={56} class="h-full shrink-0" />
	<div class="min-w-0 flex-1">
		<p class="truncate text-sm font-semibold" class:text-action={current}>
			{#if pageHref}
				<!-- Stretched over the tile: the whole tile opens the page, the play button sits above it. -->
				<a href={pageHref} class="after:absolute after:inset-0">{track.title}</a>
			{:else}
				{track.title}
			{/if}
		</p>
		<p class="truncate text-xs text-dim">{formatTrackByline(track)}</p>
	</div>
	<PlayOverlayButton
		{playing}
		size="sm"
		label="Play {track.title}"
		onclick={play}
		class={['relative z-10 shrink-0', !playing && 'opacity-0 group-hover/tile:opacity-100']}
	/>
</div>

<TrackContextMenu popup={contextMenu} {track} />
