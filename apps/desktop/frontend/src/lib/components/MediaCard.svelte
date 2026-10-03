<script lang="ts">
	import type { Item } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { player } from '#lib/player/player.svelte.ts';

	let { item }: { item: Item } = $props();

	const playing = $derived(isTrackPlaying());

	function isTrackPlaying(): boolean {
		if (!item.track) {
			return false;
		}
		return player.isPlaying(item.track);
	}

	function playTrack() {
		if (item.track) {
			player.playOrToggle(item.track);
		}
	}
</script>

<article
	class="group duration-fast flex w-44 shrink-0 flex-col gap-2 rounded-lg p-3 transition-colors hover:bg-hover"
>
	<div class="relative">
		<Cover
			thumbnails={item.thumbnails}
			size={152}
			class={[
				'w-full shadow-md',
				item.kind === 'artist' && 'rounded-full',
				item.kind !== 'artist' && 'rounded-md'
			]}
		/>
		{#if item.track}
			<PlayOverlayButton
				{playing}
				label="Play {item.title}"
				onclick={playTrack}
				class={[
					'absolute right-2 bottom-2',
					!playing && 'translate-y-2 opacity-0 group-hover:translate-y-0 group-hover:opacity-100'
				]}
			/>
		{/if}
	</div>
	<div class="min-w-0">
		<h3 class="truncate text-base font-medium">{item.title}</h3>
		<p class="line-clamp-2 text-sm text-dim">{item.subtitle}</p>
	</div>
</article>
