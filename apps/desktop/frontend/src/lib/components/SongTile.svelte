<script lang="ts">
	import type { Track } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { formatArtists } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';

	let { track }: { track: Track } = $props();

	const current = $derived(player.isCurrent(track));
	const playing = $derived(player.isPlaying(track));
</script>

<div
	class="group duration-fast flex h-16 items-center gap-3 overflow-hidden rounded-md bg-raised pr-3 transition-colors hover:bg-hover"
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
		onclick={() => player.playOrToggle(track)}
		class={['shrink-0', !playing && 'opacity-0 group-hover:opacity-100']}
	/>
</div>
