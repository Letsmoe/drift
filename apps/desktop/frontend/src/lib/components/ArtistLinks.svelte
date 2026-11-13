<!--
  Artist names, each one a link to the artist's page. Names without an ID are plain text. Text cannot be
  pressed inside a paragraph here, so this is a row of buttons and commas.
-->
<script lang="ts">
	import { Button, View } from '@neoworks/svelte-gpui/ui';
	import type { Artist } from '#lib/api/music.ts';
	import Label from '#lib/components/Label.svelte';
	import { itemHref } from '#lib/format.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { colors, type TextSize } from '#lib/theme.ts';

	let {
		artists,
		size = 'small',
		color
	}: {
		artists: Artist[];
		size?: TextSize;
		/** The text colour; the one of the parent when left out. */
		color?: string;
	} = $props();
</script>

<View style={{ direction: 'row', minWidth: 0, color }}>
	{#each artists as artist, index (index)}
		{#if index > 0}
			<Label {size} style={{ marginRight: 4 }}>,</Label>
		{/if}
		{#if artist.id}
			<Button
				hover={{ color: colors.text, underline: true }}
				onpress={() => goto(itemHref('artist', artist.id))}
			>
				<Label {size}>{artist.name}</Label>
			</Button>
		{:else}
			<Label {size}>{artist.name}</Label>
		{/if}
	{/each}
</View>
