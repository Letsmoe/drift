<script lang="ts">
	import type { Episode } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { formatDuration } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';

	let { episode, onplay }: { episode: Episode; onplay: () => void } = $props();

	const track = $derived(episode.track);
	const current = $derived(player.isCurrent(track));
	const playing = $derived(player.isPlaying(track));
	const details = $derived(
		[episode.published, formatDuration(track.durationSeconds)]
			.filter((part) => part !== '')
			.join(' • ')
	);
</script>

<div
	class="group duration-fast flex gap-4 rounded-lg p-4 transition-colors focus-within:bg-hover hover:bg-hover"
>
	<Cover thumbnails={track.thumbnails} size={96} class="size-24 shrink-0 rounded-md" />
	<div class="flex min-w-0 flex-1 flex-col gap-1">
		<h3 class={['truncate text-base font-bold', current && 'text-green']}>{track.title}</h3>
		<p class="line-clamp-2 text-sm text-dim">{episode.description}</p>
		<div class="flex items-center gap-3 pt-2">
			<PlayOverlayButton size="sm" {playing} label="Play {track.title}" onclick={onplay} />
			<span class="text-sm text-dim">{details}</span>
		</div>
	</div>
</div>
