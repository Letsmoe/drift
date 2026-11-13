<!--
  A menu at a point in the window, such as under the pointer of a right press. An outside press or Escape
  closes it; the overlay keeps it inside the window.
-->
<script lang="ts">
	import { Overlay, type KeyEvent } from '@neoworks/svelte-gpui/ui';
	import MenuList from '#lib/components/MenuList.svelte';
	import type { MenuEntry } from '#lib/menu/menuEntry.ts';

	let {
		entries,
		x,
		y,
		onclose
	}: { entries: MenuEntry[]; x: number; y: number; onclose: () => void } = $props();

	function closeOnEscape(event: KeyEvent): void {
		if (event.key === 'Escape') {
			event.preventDefault();
			onclose();
		}
	}
</script>

<Overlay
	at={[x, y]}
	side="bottom"
	align="start"
	margin={8}
	onoutside={onclose}
	onkey={closeOnEscape}
>
	<MenuList {entries} {onclose} autofocus />
</Overlay>
