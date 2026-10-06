<script lang="ts">
	import type { ItemKind, Track } from '#lib/api/music.ts';
	import Menu from '#lib/components/Menu.svelte';
	import { trackKindOf } from '#lib/format.ts';
	import type { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';
	import { buildTrackMenuEntries } from '#lib/menu/trackMenuEntries.ts';

	let { popup, track, kind }: { popup: PopupMenu; track: Track; kind?: ItemKind } = $props();

	// Tiles and recommendations don't know their kind; an episode is recognized by its show.
	const menuKind = $derived(resolveKind());

	function resolveKind(): ItemKind {
		if (kind) {
			return kind;
		}
		return trackKindOf(track);
	}
</script>

{#if popup.position}
	<Menu
		entries={buildTrackMenuEntries(track, { kind: menuKind })}
		x={popup.position.x}
		y={popup.position.y}
		onclose={() => popup.close()}
	/>
{/if}
