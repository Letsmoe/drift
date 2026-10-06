<script lang="ts">
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import type { ItemKind, Track } from '#lib/api/music.ts';
	import { library } from '#lib/library/library.svelte.ts';

	let {
		track,
		kind = 'song',
		size = 20
	}: { track: Track; kind?: ItemKind; size?: number } = $props();

	const liked = $derived(library.isLiked(track.videoId));
	const label = $derived(buildLabel());
	const weight = $derived(buildWeight());

	function buildLabel(): string {
		if (liked) {
			return 'Remove from Liked Songs';
		}
		return 'Save to Liked Songs';
	}

	function buildWeight(): 'fill' | 'regular' {
		if (liked) {
			return 'fill';
		}
		return 'regular';
	}
</script>

<button
	type="button"
	aria-label={label}
	title={label}
	aria-pressed={liked}
	onclick={() => library.toggleLike(track, kind)}
	class={[
		'duration-fast flex size-8 items-center justify-center rounded-full transition-colors hover:bg-hover',
		liked && 'text-action',
		!liked && 'text-muted hover:text-default'
	]}
>
	<HeartIcon {size} {weight} />
</button>
