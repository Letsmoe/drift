<script lang="ts">
	import MusicNoteIcon from 'phosphor-svelte/lib/MusicNoteIcon';
	import type { ClassValue } from 'svelte/elements';
	import type { Thumbnail } from '#lib/api/music.ts';
	import { pickThumbnail } from '#lib/format.ts';

	let {
		thumbnails,
		size,
		class: className = ''
	}: {
		thumbnails: Thumbnail[];
		/** Rendered width in CSS pixels; picks a thumbnail sharp enough for 2x screens. */
		size: number;
		class?: ClassValue;
	} = $props();

	const url = $derived(pickThumbnail(thumbnails, size * 2));
	// Remembers the URL that failed, so a new thumbnail gets its own chance to load.
	let failedUrl = $state<string | undefined>(undefined);
	const showImage = $derived(url !== undefined && url !== failedUrl);
</script>

{#if showImage}
	<img
		src={url}
		alt=""
		loading="lazy"
		onerror={() => (failedUrl = url)}
		class={['aspect-square object-cover', className]}
	/>
{:else}
	<div class={['flex aspect-square items-center justify-center bg-hover text-dim', className]}>
		<MusicNoteIcon size={Math.round(size / 3)} />
	</div>
{/if}
