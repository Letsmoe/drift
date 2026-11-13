<script lang="ts">
	import { window } from '@neoworks/svelte-gpui';
	import { Image, View } from '@neoworks/svelte-gpui/ui';
	import type { Thumbnail } from '#lib/api/music.ts';
	import Label from '#lib/components/Label.svelte';
	import VerticalFade from '#lib/components/VerticalFade.svelte';
	import { pickThumbnail } from '#lib/format.ts';
	import { colors, wideWindowWidth, type TextSize } from '#lib/theme.ts';

	const heroHeight = 280;

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
	<VerticalFade color="#000000" height={heroHeight} topAlpha={0.1} bottomAlpha={0.8} />
	<Label size={nameSize} weight="bold" lines={2} style={{ color: colors.white }}>{name}</Label>
</View>
