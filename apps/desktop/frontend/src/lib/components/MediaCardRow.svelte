<script lang="ts">
	import type { Item } from '#lib/api/music.ts';
	import MediaCard from '#lib/components/MediaCard.svelte';
	import ScrollRow from '#lib/components/ScrollRow.svelte';

	let { title, items }: { title: string; items: Item[] } = $props();
</script>

<section class="flex flex-col gap-2">
	<h2 class="px-3 text-2xl font-bold">{title}</h2>
	<ScrollRow>
		{#each items as item (item.id)}
			<MediaCard {item} />
		{/each}
	</ScrollRow>
</section>
