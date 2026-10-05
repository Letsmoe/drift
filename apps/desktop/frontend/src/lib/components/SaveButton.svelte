<script lang="ts">
	import CheckCircleIcon from 'phosphor-svelte/lib/CheckCircleIcon';
	import PlusCircleIcon from 'phosphor-svelte/lib/PlusCircleIcon';
	import type { Item, ItemKind } from '#lib/api/music.ts';
	import { library } from '#lib/library/library.svelte.ts';

	let { item }: { item: Item } = $props();

	const saved = $derived(library.isSaved(item.kind as ItemKind, item.id));
</script>

<button
	type="button"
	aria-label="Save {item.title} to your library"
	aria-pressed={saved}
	onclick={() => library.toggleSaved(item)}
	class={[
		'duration-fast transition-transform hover:scale-105',
		saved && 'text-action',
		!saved && 'text-dim hover:text-default'
	]}
>
	{#if saved}
		<CheckCircleIcon size={32} weight="fill" />
	{:else}
		<PlusCircleIcon size={32} />
	{/if}
</button>
