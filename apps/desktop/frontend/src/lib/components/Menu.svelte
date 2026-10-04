<script lang="ts">
	import type { Attachment } from 'svelte/attachments';
	import MenuList from '#lib/components/MenuList.svelte';
	import type { MenuEntry } from '#lib/menu/menuEntry.ts';

	let {
		entries,
		x,
		y,
		onclose
	}: { entries: MenuEntry[]; x: number; y: number; onclose: () => void } = $props();

	const viewportMargin = 8;
	let menuElement = $state<HTMLElement>();

	// Moves the menu back inside the window when it would overflow the right or bottom edge.
	const keepInViewport: Attachment<HTMLElement> = (element) => {
		const bounds = element.getBoundingClientRect();
		const maxLeft = window.innerWidth - bounds.width - viewportMargin;
		const maxTop = window.innerHeight - bounds.height - viewportMargin;
		element.style.left = `${Math.max(viewportMargin, Math.min(x, maxLeft))}px`;
		element.style.top = `${Math.max(viewportMargin, Math.min(y, maxTop))}px`;
	};

	function closeOnOutsidePointer(event: PointerEvent) {
		if (menuElement && !menuElement.contains(event.target as Node)) {
			onclose();
		}
	}

	function closeOnEscape(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			event.preventDefault();
			onclose();
		}
	}
</script>

<svelte:window
	onpointerdown={closeOnOutsidePointer}
	onkeydown={closeOnEscape}
	onblur={onclose}
	onresize={onclose}
/>

<div
	bind:this={menuElement}
	{@attach keepInViewport}
	role="presentation"
	style="left: {x}px; top: {y}px"
	oncontextmenu={(event) => event.preventDefault()}
	class="fixed z-50"
>
	<MenuList {entries} {onclose} autofocus />
</div>
