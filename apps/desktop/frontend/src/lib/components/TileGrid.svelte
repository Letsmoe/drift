<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		count,
		tile,
		collapsedRows = 2,
		columns: columnTemplate = 'repeat(auto-fill, minmax(20rem, 1fr))'
	}: {
		count: number;
		tile: Snippet<[number]>;
		collapsedRows?: number;
		/** CSS grid-template-columns; the default fits as many 20rem columns as there is room for. */
		columns?: string;
	} = $props();

	let columnCount = $state(1);
	let expanded = $state(false);

	const collapsedCount = $derived(columnCount * collapsedRows);
	const collapsible = $derived(count > collapsedCount);
	const shownCount = $derived(visibleCount());
	const toggleLabel = $derived(buildToggleLabel());

	function visibleCount(): number {
		if (expanded) {
			return count;
		}
		return Math.min(count, collapsedCount);
	}

	function buildToggleLabel(): string {
		if (expanded) {
			return 'Show less';
		}
		return `Show all ${count}`;
	}

	// The grid fills its width with as many columns as fit; read back how many that is.
	function measureColumns(grid: HTMLDivElement) {
		const observer = new ResizeObserver(() => {
			columnCount = getComputedStyle(grid).gridTemplateColumns.split(' ').length;
		});
		observer.observe(grid);
		return () => observer.disconnect();
	}
</script>

<div class="flex flex-col gap-2 px-3">
	<div class="grid gap-2" style:grid-template-columns={columnTemplate} {@attach measureColumns}>
		{#each { length: shownCount }, index (index)}
			{@render tile(index)}
		{/each}
	</div>
	{#if collapsible}
		<button
			type="button"
			aria-expanded={expanded}
			onclick={() => (expanded = !expanded)}
			class="self-start rounded-full px-3 py-1 text-sm font-semibold text-dim transition-colors hover:bg-hover hover:text-default"
		>
			{toggleLabel}
		</button>
	{/if}
</div>
