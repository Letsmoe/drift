<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { Thumbnail } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';

	let {
		label,
		title,
		thumbnails = [],
		description = '',
		cover = undefined,
		meta = undefined
	}: {
		label: string;
		title: string;
		thumbnails?: Thumbnail[];
		description?: string;
		/** Replaces the cover image, e.g. the Liked Songs tile. */
		cover?: Snippet;
		meta?: Snippet;
	} = $props();
</script>

<header class="flex items-end gap-6 pt-4 pb-6">
	<div class="size-58 shrink-0 overflow-hidden rounded-md shadow-lg">
		{#if cover}
			{@render cover()}
		{:else}
			<Cover {thumbnails} size={232} class="size-full" />
		{/if}
	</div>
	<div class="flex min-w-0 flex-col gap-2">
		<p class="text-sm font-medium">{label}</p>
		<h1 class="line-clamp-2 text-6xl leading-tight font-bold xl:text-7xl">{title}</h1>
		{#if description}
			<p class="line-clamp-2 max-w-3xl text-sm text-muted">{description}</p>
		{/if}
		{#if meta}
			<p class="flex flex-wrap items-center gap-x-1 text-sm text-muted">{@render meta()}</p>
		{/if}
	</div>
</header>
