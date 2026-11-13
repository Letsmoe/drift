<script lang="ts">
	import { window } from '@neoworks/svelte-gpui';
	import { Image, View } from '@neoworks/svelte-gpui/ui';
	import type { Thumbnail } from '#lib/api/music.ts';
	import Label from '#lib/components/Label.svelte';
	import { pickThumbnail } from '#lib/format.ts';
	import { colors, wideWindowWidth, withAlpha, type TextSize } from '#lib/theme.ts';

	const heroHeight = 280;
	// The old overlay ran from black at 10% at the top over 20% in the middle to 80% at the bottom.
	// A gradient has two stops, so it is two gradients that meet at the middle stop.
	const shadeTop = withAlpha('#000000', 0.1);
	const shadeMiddle = withAlpha('#000000', 0.2);
	const shadeBottom = withAlpha('#000000', 0.8);

	let { name, thumbnails }: { name: string; thumbnails: Thumbnail[] } = $props();

	const imageUrl = $derived(pickThumbnail(thumbnails, 1200));
	const nameSize = $derived(nameSizeFor(window.size.width));

	function nameSizeFor(windowWidth: number): TextSize {
		if (windowWidth >= wideWindowWidth) {
			return 'hero';
		}
		return 'pageTitle';
	}
</script>

<View
	style={{
		height: heroHeight,
		position: 'relative',
		justify: 'end',
		paddingX: 36,
		paddingBottom: 24,
		background: colors.raised,
		overflow: 'hidden'
	}}
>
	{#if imageUrl}
		<Image
			src={imageUrl}
			fit="cover"
			style={{
				position: 'absolute',
				top: 0,
				left: 0,
				width: 'full',
				height: heroHeight
			}}
		/>
	{/if}
	<View style={{ position: 'absolute', top: 0, left: 0, right: 0, height: heroHeight }}>
		<View
			style={{ flex: 1, background: { angle: 'to bottom', from: shadeTop, to: shadeMiddle } }}
		/>
		<View
			style={{ flex: 1, background: { angle: 'to bottom', from: shadeMiddle, to: shadeBottom } }}
		/>
	</View>
	<Label size={nameSize} weight="bold" lines={2} style={{ color: colors.white }}>{name}</Label>
</View>
