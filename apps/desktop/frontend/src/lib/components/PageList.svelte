<!--
  A page that scrolls: its content is a list of blocks (a header, a shelf, a track) of which only the ones
  near the window are built. It takes the room its parent gives it and pads each block like the old
  layout's `main` did.
-->
<script lang="ts" generics="Block">
	import type { Snippet } from 'svelte';
	import { List, View, type ScrollPosition } from '@neoworks/svelte-gpui/ui';

	let {
		items,
		row: renderRow,
		key,
		paddingX = 12,
		paddingTop = 16,
		paddingBottom = 16,
		rowGap = 0,
		estimatedRowHeight = 120,
		onscroll
	}: {
		items: Block[];
		row: Snippet<[Block, number]>;
		/** Identifies a block, so its measured height stays with it when the list changes. */
		key: (item: Block, index: number) => unknown;
		paddingX?: number;
		paddingTop?: number;
		paddingBottom?: number;
		/** Space under every block but the last. */
		rowGap?: number;
		estimatedRowHeight?: number;
		onscroll?: (position: ScrollPosition) => void;
	} = $props();

	function bottomPadding(index: number): number {
		if (index === items.length - 1) {
			return paddingBottom;
		}
		return rowGap;
	}

	function topPadding(index: number): number {
		if (index === 0) {
			return paddingTop;
		}
		return 0;
	}
</script>

<List {items} {key} {estimatedRowHeight} {onscroll} style={{ flex: 1, minHeight: 0 }}>
	{#snippet row(item, index)}
		<View
			style={{
				paddingX,
				paddingTop: topPadding(index),
				paddingBottom: bottomPadding(index)
			}}
		>
			{@render renderRow(item, index)}
		</View>
	{/snippet}
</List>
