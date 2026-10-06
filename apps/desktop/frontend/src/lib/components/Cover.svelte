<script lang="ts">
	import MusicNoteIcon from 'phosphor-svelte/lib/MusicNoteIcon';
	import type { ClassValue } from 'svelte/elements';
	import type { Thumbnail } from '#lib/api/music.ts';
	import { pickThumbnail } from '#lib/format.ts';

	let {
		thumbnails,
		size,
		aspect = 'square',
		class: className = ''
	}: {
		thumbnails: Thumbnail[];
		/** Rendered width in CSS pixels; picks a thumbnail sharp enough for 2x screens. */
		size: number;
		/** 'video' (16:9) suits episode thumbnails, which YouTube serves wide. */
		aspect?: 'square' | 'video';
		class?: ClassValue;
	} = $props();

	const url = $derived(pickThumbnail(thumbnails, size * 2));
	// Remembers the URL that failed, so a new thumbnail gets its own chance to load.
	let failedUrl = $state<string | undefined>(undefined);
	const showImage = $derived(url !== undefined && url !== failedUrl);
	const aspectClass = $derived(aspectClassFor(aspect));

	function aspectClassFor(shape: 'square' | 'video'): string {
		if (shape === 'video') {
			return 'aspect-video';
		}
		return 'aspect-square';
	}
</script>

{#if showImage}
	<img
		src={url}
		alt=""
		loading="lazy"
		onerror={() => (failedUrl = url)}
		class={[aspectClass, 'object-cover', className]}
	/>
{:else}
	<div class={['flex items-center justify-center bg-hover text-dim', aspectClass, className]}>
		<MusicNoteIcon size={Math.round(size / 3)} />
	</div>
{/if}
