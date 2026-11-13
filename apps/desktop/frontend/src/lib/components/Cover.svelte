<script lang="ts">
	import { Icon, Image, View, type Style } from '@neoworks/svelte-gpui/ui';
	import type { Thumbnail } from '#lib/api/music.ts';
	import { pickThumbnail } from '#lib/format.ts';
	import { musicNote } from '#lib/icons.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import { colors } from '#lib/theme.ts';

	let {
		thumbnails,
		size,
		aspect = 'square',
		fill = false,
		style,
		onload
	}: {
		/** Called when the image has loaded, which happens once it is near the window. */
		onload?: () => void;
		thumbnails: Thumbnail[];
		/** Rendered width in pixels; picks a thumbnail sharp enough for 2x screens. */
		size: number;
		/** 'video' (16:9) suits episode thumbnails, which YouTube serves wide. */
		aspect?: 'square' | 'video';
		/** Fills the parent instead of keeping an aspect, e.g. as a backdrop; aspect is ignored. */
		fill?: boolean;
		/** Overrides the box: a fixed width and height, a radius, a shadow. */
		style?: Style;
	} = $props();

	const url = $derived(pickThumbnail(thumbnails, size * 2));
	// Remembers the URL that failed, so a new thumbnail gets its own chance to load.
	let failedUrl = $state<string | undefined>(undefined);
	const showImage = $derived(url !== undefined && url !== failedUrl);
	const box = $derived(boxStyle());
	// Only an image that rounds itself can be a circle, and it cannot also be cropped (an artist's photo is square).
	const circular = $derived(typeof box.radius === 'number' && box.radius >= size / 2);

	function boxStyle(): Style {
		if (fill) {
			return { ...style, size: 'full' };
		}
		if (aspect === 'video') {
			return { width: 'full', aspectRatio: 16 / 9, ...style };
		}
		return { width: 'full', aspectRatio: 1, ...style };
	}
</script>

{#if showImage && url !== undefined}
	{#if circular}
		<Image
			src={url}
			fit="contain"
			style={{ ...box, background: 'rgba(0, 0, 0, 0.4)' }}
			{onload}
			onerror={() => (failedUrl = url)}
		/>
	{:else}
		<!-- A cropping image is not clipped to its own box, so a box around it clips it. -->
		<View style={{ ...box, overflow: 'hidden', background: 'rgba(0, 0, 0, 0.4)' }}>
			<Image
				src={url}
				fit="cover"
				style={{ size: 'full' }}
				{onload}
				onerror={() => (failedUrl = url)}
			/>
		</View>
	{/if}
{:else}
	<View
		style={{
			...box,
			background: colors.hover,
			color: appearance.dim,
			align: 'center',
			justify: 'center'
		}}
	>
		<Icon svg={musicNote} size={Math.round(size / 3)} />
	</View>
{/if}
