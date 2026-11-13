<!--
  A box that opens a popup menu on a right press and reports whether the pointer is over it. Left presses
  pass through to what is inside.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import { View, type PointerEvent, type Style } from '@neoworks/svelte-gpui/ui';
	import type { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';

	let {
		popup,
		onhover,
		onpress,
		style,
		children
	}: {
		popup?: PopupMenu;
		onhover?: (hovered: boolean) => void;
		/** A left press on the box itself, not on a button inside it. */
		onpress?: (event: PointerEvent) => void;
		style?: Style;
		children: Snippet;
	} = $props();

	function openMenu(event: PointerEvent): void {
		popup?.openAtPointer(event);
	}
</script>

<View {style} {onhover} {onpress} buttons={['right']} ondown={openMenu}>
	{@render children()}
</View>
