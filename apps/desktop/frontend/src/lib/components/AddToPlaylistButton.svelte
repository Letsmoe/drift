<script lang="ts">
	import PlusCircleIcon from 'phosphor-svelte/lib/PlusCircleIcon';
	import type { Track } from '#lib/api/music.ts';
	import Menu from '#lib/components/Menu.svelte';
	import { playlistEntries } from '#lib/menu/trackMenuEntries.ts';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';

	let { track, large = false }: { track: Track; large?: boolean } = $props();

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
	class={[
		'duration-fast flex items-center justify-center rounded-full text-muted transition-colors hover:bg-hover hover:text-default',
		large && 'p-2',
		!large && 'size-8'
	]}
>
	{#if large}
		<PlusCircleIcon size={24} />
	{:else}
		<PlusCircleIcon size={20} />
	{/if}
</button>

{#if popup.position}
	<Menu
		entries={playlistEntries(track)}
		x={popup.position.x}
		y={popup.position.y}
		onclose={() => popup.close()}
	/>
{/if}
