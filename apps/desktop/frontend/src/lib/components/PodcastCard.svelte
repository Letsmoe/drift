<script lang="ts">
	import type { Item } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import { itemHref } from '#lib/format.ts';

	let { item }: { item: Item } = $props();
</script>

<a
	href={itemHref('podcast', item.id)}
	class="duration-fast flex gap-3 rounded-lg bg-raised p-3 transition-colors hover:bg-hover"
>
	<Cover thumbnails={item.thumbnails} size={96} class="size-24 shrink-0 rounded-md shadow-md" />
	<div class="flex min-w-0 flex-1 flex-col gap-0.5">
		<h3 class="line-clamp-2 text-sm leading-snug font-semibold">{item.title}</h3>
		<p class="truncate text-xs text-muted">{item.subtitle}</p>
		{#if item.description}
			<p class="line-clamp-3 pt-1 text-xs text-dim">{item.description}</p>
		{/if}
	</div>
</a>
