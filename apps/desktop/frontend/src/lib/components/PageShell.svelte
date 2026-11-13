<!--
  The frame of an album, playlist or podcast page: a tinted wash under the header that fades out, with the
  page's own scrolling content over it. The wash stays at the top of the pane while the content scrolls.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import { View } from '@neoworks/svelte-gpui/ui';
	import type { Thumbnail } from '#lib/api/music.ts';
	import { gradientColorFor, gradientColorFromCover } from '#lib/pageColor.ts';
	import { CoverPalette } from '#lib/palette/coverPalette.svelte.ts';
	import { radii } from '#lib/theme.ts';

	const washHeight = 416;
	// The cover's colours arrive a moment after the page; the wash eases into them.
	const washTransitionMs = 400;

	let {
		seed,
		cover = [],
		children
	}: { seed: string; cover?: Thumbnail[]; children: Snippet } = $props();

	const coverPalette = new CoverPalette();

	$effect(() => {
		coverPalette.load(cover);
	});

	const color = $derived(washColor());

	function washColor(): string {
		const vibrant = coverPalette.colors?.vibrant;
		if (vibrant === undefined) {
			return gradientColorFor(seed);
		}
		const fromCover = gradientColorFromCover(vibrant);
		if (fromCover === undefined) {
			return gradientColorFor(seed);
		}
		return fromCover;
	}
</script>

<View style={{ flex: 1, minHeight: 0, position: 'relative' }}>
	<!-- The pane clips with square edges, so the wash rounds its own top corners to match the pane's. -->
	<View
		style={{
			position: 'absolute',
			top: 0,
			left: 0,
			right: 0,
			height: washHeight,
			radius: [radii.extraLarge, radii.extraLarge, 0, 0],
			background: { angle: 'to bottom', from: color, to: 'transparent' },
			transition: washTransitionMs
		}}
	/>
	{@render children()}
</View>
