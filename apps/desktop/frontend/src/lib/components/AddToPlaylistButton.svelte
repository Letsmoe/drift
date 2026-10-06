<script lang="ts">
	import PlusCircleIcon from 'phosphor-svelte/lib/PlusCircleIcon';
	import type { Track } from '#lib/api/music.ts';
	import Menu from '#lib/components/Menu.svelte';
	import { playlistEntries } from '#lib/menu/trackMenuEntries.ts';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';

	let { track }: { track: Track } = $props();

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
	aria-label="Add {track.title} to a playlist"
	title="Add to playlist"
	aria-haspopup="menu"
	aria-expanded={popup.isOpen}
	onclick={toggle}
	class="duration-fast flex size-8 items-center justify-center rounded-full text-muted transition-colors hover:bg-hover hover:text-default"
>
	<PlusCircleIcon size={20} />
</button>

{#if popup.position}
	<Menu
		entries={playlistEntries(track)}
		x={popup.position.x}
		y={popup.position.y}
		onclose={() => popup.close()}
	/>
{/if}
