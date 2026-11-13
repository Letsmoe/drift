<script lang="ts">
	import { Popover, type Style } from '@neoworks/svelte-gpui/ui';
	import IconAction from '#lib/components/IconAction.svelte';
	import MenuList from '#lib/components/MenuList.svelte';
	import { appearance } from '#lib/appearance.svelte.ts';
	import { dotsThreeBold } from '#lib/icons.ts';
	import type { MenuEntry } from '#lib/menu/menuEntry.ts';
	import { colors } from '#lib/theme.ts';

	let {
		buildEntries,
		size = 32,
		boxSize = 48,
		open = $bindable(false),
		style
	}: {
		buildEntries: () => MenuEntry[];
		/** The icon's size. */
		size?: number;
		/** The pressable area's size. */
		boxSize?: number;
		open?: boolean;
		/** Layout and opacity of the button, e.g. to hide it until its row is hovered. */
		style?: Style;
	} = $props();
</script>

<Popover bind:open style={{}} focusPanel={false} side="bottom" align="start" offset={4}>
	{#snippet trigger({ toggle })}
		<IconAction
			icon={dotsThreeBold}
			iconSize={size}
			{boxSize}
			color={appearance.dim}
			hoverColor={colors.text}
			{style}
			onpress={toggle}
		/>
	{/snippet}
	{#snippet children({ close })}
		<MenuList entries={buildEntries()} onclose={() => close('select')} autofocus />
	{/snippet}
</Popover>
