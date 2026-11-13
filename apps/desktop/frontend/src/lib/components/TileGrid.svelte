<!--
  Tiles in as many equal columns as fit, collapsed to a couple of rows with a "Show all" button. The old
  grid was CSS; here the column count is worked out from the measured width and every tile gets its width.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Button, View } from '@neoworks/svelte-gpui/ui';
	import Label from '#lib/components/Label.svelte';
	import { appearance } from '#lib/appearance.svelte.ts';
	import { colors, gutter, hoverTransition, radii } from '#lib/theme.ts';

	const gap = 8;
	const minimumTileWidth = 320;

	let {
		count,
		tile,
		collapsedRows = 2,
		fixedColumns
	}: {
		count: number;
		tile: Snippet<[number]>;
		collapsedRows?: number;
		/** A set number of columns instead of as many 20rem columns as fit. */
		fixedColumns?: number;
	} = $props();

	let width = $state(0);
	let expanded = $state(false);

	const columnCount = $derived(columnCountFor(width));
	const tileWidth = $derived(Math.floor((width - gap * (columnCount - 1)) / columnCount));
	const collapsedCount = $derived(columnCount * collapsedRows);
	const collapsible = $derived(count > collapsedCount);
	const shownCount = $derived(shownCountFor());
	const toggleLabel = $derived(toggleLabelFor());

	function columnCountFor(measuredWidth: number): number {
		if (fixedColumns !== undefined) {
			return fixedColumns;
		}
		return Math.max(1, Math.floor((measuredWidth + gap) / (minimumTileWidth + gap)));
	}

	function shownCountFor(): number {
		if (expanded) {
			return count;
		}
		return Math.min(count, collapsedCount);
	}

	function toggleLabelFor(): string {
		if (expanded) {
			return 'Show less';
		}
		return `Show all ${count}`;
	}
</script>

<View style={{ gap: 8, paddingX: gutter.narrow }}>
	<View
		onresize={(layout) => (width = layout.width)}
		style={{ direction: 'row', wrap: true, gap, align: 'stretch' }}
	>
		{#if width > 0}
			{#each { length: shownCount }, index (index)}
				<View style={{ width: tileWidth }}>
					{@render tile(index)}
				</View>
			{/each}
		{/if}
	</View>
	{#if collapsible}
		<Button
			style={{
				alignSelf: 'start',
				paddingX: 12,
				paddingY: 4,
				radius: radii.full,
				color: appearance.dim,
				transition: hoverTransition
			}}
			hover={{ background: colors.hover, color: colors.text }}
			onpress={() => (expanded = !expanded)}
		>
			<Label size="small" weight="semibold">{toggleLabel}</Label>
		</Button>
	{/if}
</View>
