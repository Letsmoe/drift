<script lang="ts">
	import type { Track } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { formatArtists } from '#lib/format.ts';
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

	const current = $derived(player.isCurrent(track));
	const playing = $derived(player.isPlaying(track));

	function play(): void {
		onplay?.();
		player.playOrToggle(track, tracks, index, context);
	}

	function reportVisible(): void {
		onvisible?.();
	}
</script>

<div
	class="group duration-fast flex h-16 items-center gap-3 overflow-hidden rounded-md bg-raised pr-3 transition-colors hover:bg-hover"
	{@attach whenVisible(reportVisible)}
>
	<Cover thumbnails={track.thumbnails} size={64} class="h-full shrink-0" />
	<div class="min-w-0 flex-1">
		<p class="truncate text-base font-semibold" class:text-green={current}>{track.title}</p>
		<p class="truncate text-sm text-dim">{formatArtists(track.artists)}</p>
	</div>
	<PlayOverlayButton
		{playing}
		size="sm"
		label="Play {track.title}"
		onclick={play}
		class={['shrink-0', !playing && 'opacity-0 group-hover:opacity-100']}
	/>
</div>
