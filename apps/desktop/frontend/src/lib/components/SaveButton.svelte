<script lang="ts">
	import type { Item, ItemKind } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import IconAction from '#lib/components/IconAction.svelte';
	import { checkCircleFill, plusCircle } from '#lib/icons.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { colors } from '#lib/theme.ts';

	let { item }: { item: Item } = $props();

	const saved = $derived(library.isSaved(item.kind as ItemKind, item.id));
	const icon = $derived(iconFor(saved));
	const color = $derived(colorFor(saved));

	function iconFor(isSaved: boolean): string {
		if (isSaved) {
			return checkCircleFill;
		}
		return plusCircle;
	}

	function colorFor(isSaved: boolean): string {
		if (isSaved) {
			return appearance.accent;
		}
		return appearance.dim;
	}
</script>

<IconAction
	{icon}
	iconSize={32}
	{color}
	hoverColor={colors.text}
	onpress={() => library.toggleSaved(item)}
/>
