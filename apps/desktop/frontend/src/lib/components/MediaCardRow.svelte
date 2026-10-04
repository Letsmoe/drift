<script lang="ts">
	import { FloatingScrollbar } from '@neoworks-dev/ui';
	import type { Item } from '#lib/api/music.ts';
	import MediaCard from '#lib/components/MediaCard.svelte';

	let { title, items }: { title: string; items: Item[] } = $props();
</script>

<section class="flex flex-col gap-2">
	<h2 class="px-3 text-2xl font-bold">{title}</h2>
	<FloatingScrollbar axis="horizontal">
		<div class="flex">
			{#each items as item (item.id)}
				<MediaCard {item} />
			{/each}
		</div>
	</FloatingScrollbar>
</section>
