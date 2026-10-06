<script lang="ts">
	import type { ItemKind, Track } from '#lib/api/music.ts';
	import Menu from '#lib/components/Menu.svelte';
	import type { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';
	import { buildTrackMenuEntries } from '#lib/menu/trackMenuEntries.ts';

	let {
		popup,
		track,
		kind = 'song'
	}: { popup: PopupMenu; track: Track; kind?: ItemKind } = $props();
</script>

{#if popup.position}
	<Menu
		entries={buildTrackMenuEntries(track, { kind })}
		x={popup.position.x}
		y={popup.position.y}
		onclose={() => popup.close()}
	/>
{/if}
