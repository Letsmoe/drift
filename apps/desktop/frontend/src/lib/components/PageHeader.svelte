<script lang="ts">
	import type { Snippet } from 'svelte';
	import { window } from '@neoworks/svelte-gpui';
	import { View } from '@neoworks/svelte-gpui/ui';
	import type { Thumbnail } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Cover from '#lib/components/Cover.svelte';
	import Label from '#lib/components/Label.svelte';
	import { radii, wideWindowWidth, type TextSize } from '#lib/theme.ts';

	const coverSide = 232;

	let {
		label,
		title,
		thumbnails = [],
		description = '',
		cover = undefined,
		meta = undefined
	}: {
		label: string;
		title: string;
		thumbnails?: Thumbnail[];
		description?: string;
		/** Replaces the cover image, e.g. the Liked Songs tile. */
		cover?: Snippet;
		meta?: Snippet;
	} = $props();

	const titleSize = $derived(titleSizeFor(window.size.width));

	function titleSizeFor(windowWidth: number): TextSize {
		if (windowWidth >= wideWindowWidth) {
			return 'pageTitleWide';
		}
		return 'pageTitle';
	}
</script>

<View style={{ direction: 'row', align: 'end', gap: 24, paddingTop: 16, paddingBottom: 24 }}>
	<View
		style={{
			size: coverSide,
			shrink: 0,
			radius: radii.medium,
			overflow: 'hidden',
			shadow: { y: 12, blur: 32, color: 'rgba(0, 0, 0, 0.55)' }
		}}
	>
		{#if cover}
			{@render cover()}
		{:else}
			<Cover {thumbnails} size={coverSide} style={{ size: coverSide }} />
		{/if}
	</View>
	<View style={{ flex: 1, minWidth: 0, gap: 8 }}>
		<Label size="small" weight="medium">{label}</Label>
		<Label size={titleSize} weight="bold" lines={2}>{title}</Label>
		{#if description}
			<Label size="small" tone="muted" lines={2} style={{ maxWidth: 768 }}>{description}</Label>
		{/if}
		{#if meta}
			<View
				style={{
					direction: 'row',
					wrap: true,
					align: 'center',
					columnGap: 4,
					color: appearance.muted
				}}
			>
				{@render meta()}
			</View>
		{/if}
	</View>
</View>
