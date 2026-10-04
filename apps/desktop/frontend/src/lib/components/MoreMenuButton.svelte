<script lang="ts">
	import DotsThreeIcon from 'phosphor-svelte/lib/DotsThreeIcon';
	import Menu from '#lib/components/Menu.svelte';
	import type { MenuEntry } from '#lib/menu/menuEntry.ts';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';

	let {
		buildEntries,
		label = 'More options',
		size = 32
	}: { buildEntries: () => MenuEntry[]; label?: string; size?: number } = $props();

	const popup = new PopupMenu();

	function toggle(event: MouseEvent) {
		event.stopPropagation();
		if (popup.isOpen) {
			popup.close();
			return;
		}
		popup.openBelow(event.currentTarget as HTMLElement);
	}
</script>

<button
	type="button"
	aria-label={label}
	aria-haspopup="menu"
	aria-expanded={popup.isOpen}
	onclick={toggle}
	class="duration-fast flex size-12 items-center justify-center rounded-full text-dim transition-colors hover:text-default focus-visible:text-default"
>
	<DotsThreeIcon {size} weight="bold" />
</button>

{#if popup.position}
	<Menu
		entries={buildEntries()}
		x={popup.position.x}
		y={popup.position.y}
		onclose={() => popup.close()}
	/>
{/if}
