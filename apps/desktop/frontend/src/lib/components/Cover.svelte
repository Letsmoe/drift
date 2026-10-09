<script lang="ts">
	import MusicNoteIcon from 'phosphor-svelte/lib/MusicNoteIcon';
	import type { ClassValue } from 'svelte/elements';
	import type { Thumbnail } from '#lib/api/music.ts';
	import { pickThumbnail } from '#lib/format.ts';

	let {
		thumbnails,
		size,
		aspect = 'square',
		fill = false,
		class: className = ''
	}: {
		thumbnails: Thumbnail[];
		/** Rendered width in CSS pixels; picks a thumbnail sharp enough for 2x screens. */
		size: number;
		/** 'video' (16:9) suits episode thumbnails, which YouTube serves wide. */
		aspect?: 'square' | 'video';
		/** Fills the parent and crops instead of letterboxing, e.g. as a backdrop; aspect is ignored. */
		fill?: boolean;
		class?: ClassValue;
	} = $props();

	const url = $derived(pickThumbnail(thumbnails, size * 2));
	// Remembers the URL that failed, so a new thumbnail gets its own chance to load.
	let failedUrl = $state<string | undefined>(undefined);
	const showImage = $derived(url !== undefined && url !== failedUrl);
	const shapeClass = $derived(shapeClassFor(aspect));
	const fitClass = $derived(fitClassFor());

	function shapeClassFor(shape: 'square' | 'video'): string {
		if (fill) {
			return 'size-full';
		}
		if (shape === 'video') {
			return 'aspect-video';
		}
		return 'aspect-square';
	}

	function fitClassFor(): string {
		if (fill) {
			return 'object-cover';
		}
		return 'bg-black/40 object-contain';
	}
</script>

{#if showImage}
	<img
		src={url}
		alt=""
		loading="lazy"
		onerror={() => (failedUrl = url)}
		class={[shapeClass, fitClass, className]}
	/>
{:else}
	<div class={['flex items-center justify-center bg-hover text-dim', shapeClass, className]}>
		<MusicNoteIcon size={Math.round(size / 3)} />
	</div>
{/if}
