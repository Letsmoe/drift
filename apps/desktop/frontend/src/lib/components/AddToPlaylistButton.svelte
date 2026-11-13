<script lang="ts">
	import { Menu } from '@neoworks/svelte-gpui-ui';
	import type { Track } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import IconAction from '#lib/components/IconAction.svelte';
	import { plusCircle } from '#lib/icons.ts';
	import { toKitMenuItems } from '#lib/menu/kitMenuItems.ts';
	import { playlistEntries } from '#lib/menu/trackMenuEntries.ts';
	import { colors } from '#lib/theme.ts';

	let { track, large = false }: { track: Track; large?: boolean } = $props();

	let open = $state(false);
	const iconSize = $derived(iconSizeFor());
	const boxSize = $derived(boxSizeFor());
	const items = $derived(toKitMenuItems(playlistEntries(track)));

	function iconSizeFor(): number {
		if (large) {
			return 24;
		}
		return 20;
	}

	function boxSizeFor(): number {
		if (large) {
			return 40;
		}
		return 32;
	}
</script>

<Menu bind:open {items} placement="bottom-start">
	{#snippet trigger({ onpress })}
		<IconAction
			icon={plusCircle}
			{iconSize}
			{boxSize}
			color={appearance.muted}
			hoverColor={colors.text}
			hoverBackground={colors.hover}
			{onpress}
		/>
	{/snippet}
</Menu>
