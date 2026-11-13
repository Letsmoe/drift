<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Scroll, View, type ScrollPosition } from '@neoworks/svelte-gpui/ui';
	import IconAction from '#lib/components/IconAction.svelte';
	import { caretLeftBold, caretRightBold } from '#lib/icons.ts';
	import { colors } from '#lib/theme.ts';

	// Pixels of slack so rounding at either end doesn't leave an arrow that does nothing.
	const edgeTolerance = 4;
	// Each press scrolls this share of the visible width, keeping a card of context.
	const pageShare = 0.85;
	// The arrows sit a third of the way down the row, over the covers.
	const arrowTop = 78;

	let { children }: { children: Snippet } = $props();

	let scroller: ReturnType<typeof Scroll> | undefined = $state();
	let position = $state<ScrollPosition>({ x: 0, y: 0, maxX: 0, maxY: 0 });
	let visibleWidth = $state(0);
	let hovered = $state(false);

	const canScrollBack = $derived(position.x > edgeTolerance);
	const canScrollForward = $derived(position.maxX - position.x > edgeTolerance);

	function scrollByPage(direction: 1 | -1): void {
		scroller?.scrollBy(direction * visibleWidth * pageShare);
	}
</script>

<View
	style={{ position: 'relative' }}
	onhover={(isHovered) => (hovered = isHovered)}
	onresize={(layout) => (visibleWidth = layout.width)}
>
	<Scroll bind:this={scroller} horizontal onscroll={(next) => (position = next)}>
		<View style={{ direction: 'row' }}>
			{@render children()}
		</View>
	</Scroll>
	{#if hovered && canScrollBack}
		<IconAction
			icon={caretLeftBold}
			iconSize={18}
			boxSize={36}
			color={colors.text}
			hoverBackground={colors.hover}
			style={{ position: 'absolute', left: 4, top: arrowTop, background: colors.elevated }}
			onpress={() => scrollByPage(-1)}
		/>
	{/if}
	{#if hovered && canScrollForward}
		<IconAction
			icon={caretRightBold}
			iconSize={18}
			boxSize={36}
			color={colors.text}
			hoverBackground={colors.hover}
			style={{ position: 'absolute', right: 4, top: arrowTop, background: colors.elevated }}
			onpress={() => scrollByPage(1)}
		/>
	{/if}
</View>
