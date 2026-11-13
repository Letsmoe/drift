<script lang="ts">
	import { Popover } from '@neoworks/svelte-gpui/ui';
	import type { Track } from '#lib/api/music.ts';
	import IconAction from '#lib/components/IconAction.svelte';
	import MenuList from '#lib/components/MenuList.svelte';
	import { plusCircle } from '#lib/icons.ts';
	import { playlistEntries } from '#lib/menu/trackMenuEntries.ts';
	import { colors } from '#lib/theme.ts';
	import { appearance } from '#lib/appearance.svelte.ts';

	let { track, large = false }: { track: Track; large?: boolean } = $props();

	let open = $state(false);
	const iconSize = $derived(iconSizeFor());
	const boxSize = $derived(boxSizeFor());

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

<Popover bind:open style={{}} focusPanel={false} side="bottom" align="start" offset={4}>
	{#snippet trigger({ toggle })}
		<IconAction
			icon={plusCircle}
			{iconSize}
			{boxSize}
			color={appearance.muted}
			hoverColor={colors.text}
			hoverBackground={colors.hover}
			onpress={toggle}
		/>
	{/snippet}
	{#snippet children({ close })}
		<MenuList entries={playlistEntries(track)} onclose={() => close('select')} autofocus />
	{/snippet}
</Popover>
