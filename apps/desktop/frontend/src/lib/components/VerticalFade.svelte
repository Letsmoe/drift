<!--
  A colour that fades out downwards, drawn as stacked bands because `Style` has no gradients. The bands are
  absolutely positioned, so put it first in a relatively positioned parent and the content over it.
-->
<script lang="ts">
	import { View, type Color } from '@neoworks/svelte-gpui/ui';
	import { withAlpha } from '#lib/theme.ts';

	const bandCount = 40;

	let {
		color,
		height,
		topAlpha = 1,
		bottomAlpha = 0
	}: {
		/** A #rrggbb colour. */
		color: Color;
		height: number;
		/** Opacity at the top edge, from 0 to 1. */
		topAlpha?: number;
		/** Opacity at the bottom edge. */
		bottomAlpha?: number;
	} = $props();

	const bandHeight = $derived(height / bandCount);
	const bandColors = $derived(buildBandColors());

	function buildBandColors(): Color[] {
		const built: Color[] = [];
		for (let band = 0; band < bandCount; band += 1) {
			const share = (band + 0.5) / bandCount;
			built.push(withAlpha(color, topAlpha + (bottomAlpha - topAlpha) * share));
		}
		return built;
	}
</script>

<View style={{ position: 'absolute', top: 0, left: 0, right: 0, height }}>
	{#each bandColors as bandColor, band (band)}
		<View style={{ height: bandHeight, background: bandColor, shrink: 0 }} />
	{/each}
</View>
