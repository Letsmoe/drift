<script lang="ts">
	import type { Track } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { formatArtists } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { whenVisible } from '#lib/whenVisible.ts';

	let { track, onplay, onvisible }: { track: Track; onplay: () => void; onvisible?: () => void } =
		$props();

	const playing = $derived(player.isPlaying(track));

	function reportVisible(): void {
		onvisible?.();
	}
</script>

<div
	class="group duration-fast flex w-40 shrink-0 flex-col gap-2 rounded-lg p-3 transition-colors hover:bg-raised"
	{@attach whenVisible(reportVisible)}
>
	<div class="relative">
		<Cover thumbnails={track.thumbnails} size={136} class="w-full rounded-md" />
		<PlayOverlayButton
			{playing}
			size="md"
			label="Play {track.title}"
			onclick={onplay}
			class={[
				'absolute right-2 bottom-2',
				!playing && 'translate-y-1 opacity-0 group-hover:translate-y-0 group-hover:opacity-100'
			]}
		/>
	</div>
	<div class="min-w-0">
		<p class="truncate text-sm font-semibold">{track.title}</p>
		<p class="truncate text-xs text-dim">{formatArtists(track.artists)}</p>
	</div>
</div>
