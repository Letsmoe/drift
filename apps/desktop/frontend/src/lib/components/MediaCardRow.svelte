<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import type { Item } from '#lib/api/music.ts';
	import Label from '#lib/components/Label.svelte';
	import MediaCard from '#lib/components/MediaCard.svelte';
	import ScrollRow from '#lib/components/ScrollRow.svelte';
	import { gutter } from '#lib/theme.ts';

	let { title, items }: { title: string; items: Item[] } = $props();
</script>

<View style={{ gap: 8 }}>
	<Label size="heading" weight="bold" style={{ paddingX: gutter.narrow }}>{title}</Label>
	<ScrollRow>
		{#each items as item (item.id)}
			<MediaCard {item} />
		{/each}
	</ScrollRow>
</View>
