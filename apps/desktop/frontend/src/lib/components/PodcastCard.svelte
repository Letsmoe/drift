<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import type { Item } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import Label from '#lib/components/Label.svelte';
	import { itemHref } from '#lib/format.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	let { item }: { item: Item } = $props();
</script>

<View
	style={{
		cursor: 'pointer',
		flex: 1,
		direction: 'row',
		align: 'start',
		gap: 12,
		padding: 12,
		radius: radii.large,
		background: colors.raised,
		transition: hoverTransition
	}}
	hover={{ background: colors.hover }}
	onpress={() => goto(itemHref('podcast', item.id))}
>
	<Cover
		thumbnails={item.thumbnails}
		size={96}
		style={{
			width: 96,
			height: 96,
			shrink: 0,
			radius: radii.medium,
			shadow: { y: 4, blur: 12, color: 'rgba(0, 0, 0, 0.45)' }
		}}
	/>
	<View style={{ flex: 1, minWidth: 0, gap: 2 }}>
		<Label size="small" weight="semibold" lines={2}>{item.title}</Label>
		<Label size="tiny" tone="muted" truncate>{item.subtitle}</Label>
		{#if item.description}
			<Label size="tiny" tone="dim" lines={3} style={{ paddingTop: 4 }}>{item.description}</Label>
		{/if}
	</View>
</View>
