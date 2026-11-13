<!--
  A colour that fades out downwards (or in), drawn as stacked layers because `Style` has no gradients. The
  layers are absolutely positioned, so put it first in a relatively positioned parent and the content over it.

  The layers overlap instead of sitting side by side: a layer that ends at a fractional pixel adds only one
  small step there, where two neighbouring bands would each cover part of that pixel and leave a faint line.
-->
<script lang="ts">
	import { View, type Color } from '@neoworks/svelte-gpui/ui';
	import { withAlpha } from '#lib/theme.ts';

	const layerCount = 24;
	// Below this, what the layers above already cover is taken as opaque.
	const opaqueEpsilon = 0.0001;

	interface FadeLayer {
		top: number;
		height: number;
		color: Color;
	}

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

	const layers = $derived(buildLayers());

	// The opacity wanted in the middle of each of `layerCount` equal slices, top to bottom.
	function slicedAlphas(): number[] {
		const alphas: number[] = [];
		for (let slice = 0; slice < layerCount; slice += 1) {
			const share = (slice + 0.5) / layerCount;
			alphas.push(topAlpha + (bottomAlpha - topAlpha) * share);
		}
		return alphas;
	}

	// What a layer must add to `below` (the opacity already there) to reach `wanted`.
	function addedAlpha(wanted: number, below: number): number {
		if (1 - below < opaqueEpsilon) {
			return 0;
		}
		return Math.max(0, 1 - (1 - wanted) / (1 - below));
	}

	// Slice k is covered by the layers k to the last, each reaching from the top down to its own slice.
	function fadeOutLayer(alphas: number[], slice: number): FadeLayer {
		let below = 0;
		if (slice < layerCount - 1) {
			below = alphas[slice + 1];
		}
		return {
			top: 0,
			height: (height / layerCount) * (slice + 1),
			color: withAlpha(color, addedAlpha(alphas[slice], below))
		};
	}

	// Slice k is covered by the layers 0 to k, each reaching from its own slice down to the bottom.
	function fadeInLayer(alphas: number[], slice: number): FadeLayer {
		let above = 0;
		if (slice > 0) {
			above = alphas[slice - 1];
		}
		const top = (height / layerCount) * slice;
		return {
			top,
			height: height - top,
			color: withAlpha(color, addedAlpha(alphas[slice], above))
		};
	}

	function buildLayers(): FadeLayer[] {
		const alphas = slicedAlphas();
		const fadesOut = bottomAlpha <= topAlpha;
		const built: FadeLayer[] = [];
		for (let slice = 0; slice < layerCount; slice += 1) {
			if (fadesOut) {
				built.push(fadeOutLayer(alphas, slice));
				continue;
			}
			built.push(fadeInLayer(alphas, slice));
		}
		return built;
	}
</script>

<View style={{ position: 'absolute', top: 0, left: 0, right: 0, height }}>
	{#each layers as layer, index (index)}
		<View
			style={{
				position: 'absolute',
				top: layer.top,
				left: 0,
				right: 0,
				height: layer.height,
				background: layer.color
			}}
		/>
	{/each}
</View>
