<!--
  A modal dialog over a dimmed window: a card with a title and the children. A press on the dimmed area or
  Escape closes it. Mount it while it should show; `onclose` is told when it closes itself.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import { window } from '@neoworks/svelte-gpui';
	import { Popover, Scroll, View } from '@neoworks/svelte-gpui/ui';
	import Label from '#lib/components/Label.svelte';
	import { colors, radii } from '#lib/theme.ts';

	const narrowWidth = 448;
	const wideWidth = 672;
	const layerLayer = 1;

	let {
		title,
		onclose,
		wide = false,
		children: content
	}: { title: string; onclose: () => void; wide?: boolean; children: Snippet } = $props();

	let open = $state(true);
	const cardWidth = $derived(cardWidthFor(wide));

	function cardWidthFor(isWide: boolean): number {
		if (isWide) {
			return wideWidth;
		}
		return narrowWidth;
	}
</script>

<Popover
	bind:open
	at={[0, 0]}
	side="bottom"
	align="start"
	offset={0}
	flip={false}
	shift={false}
	margin={0}
	layer={layerLayer}
	closeOnOutside={false}
	style={{
		width: window.size.width,
		height: window.size.height,
		align: 'center',
		justify: 'center',
		padding: 16
	}}
	onclose={() => onclose()}
>
	{#snippet children({ close })}
		<View
			style={{ position: 'absolute', inset: 0, background: colors.scrim }}
			onpress={() => close('outside')}
		/>
		<View
			style={{
				width: 'full',
				maxWidth: cardWidth,
				maxHeight: 'full',
				padding: 24,
				background: colors.elevated,
				radius: radii.large,
				color: colors.text,
				shadow: { y: 12, blur: 32, color: 'rgba(0, 0, 0, 0.55)' }
			}}
		>
			<Label size="heading" weight="bold" style={{ paddingBottom: 16 }}>{title}</Label>
			<Scroll style={{ shrink: 1, minHeight: 0 }}>
				{@render content()}
			</Scroll>
		</View>
	{/snippet}
</Popover>
