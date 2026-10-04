<script lang="ts">
	import type { Item, ItemKind } from '#lib/api/music.ts';
	import { library } from '#lib/library/library.svelte.ts';

	let { item }: { item: Item } = $props();

	const following = $derived(library.isSaved(item.kind as ItemKind, item.id));
</script>

<button
	type="button"
	aria-pressed={following}
	onclick={() => library.toggleSaved(item)}
	class="duration-fast h-8 rounded-full border border-line-strong px-4 text-sm font-semibold transition-[transform,border-color] hover:scale-105 hover:border-default"
>
	{#if following}Following{:else}Follow{/if}
</button>
