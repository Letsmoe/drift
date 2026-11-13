<!--
  Artist names, each one a link to the artist's page. Names without an ID are plain text. Text cannot be
  pressed inside a paragraph here, so this is a row of buttons and commas.

  The names stay on one line at their full width. A line too long for its box is cut off at the edge and, while
  the pointer is over it, glides through to its end and back (`Marquee`), instead of squeezing every name.
-->
<script lang="ts">
	import { Button, View } from '@neoworks/svelte-gpui/ui';
	import type { Artist } from '#lib/api/music.ts';
	import Label from '#lib/components/Label.svelte';
	import { itemHref } from '#lib/format.ts';
	import { Marquee } from '#lib/motion/marquee.svelte.ts';
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

	const marquee = new Marquee();
	const oneLine = { nowrap: true } as const;

	function followPointer(hovered: boolean): void {
		if (hovered) {
			marquee.start();
			return;
		}
		marquee.stop();
	}
</script>

<View
	style={{ direction: 'row', minWidth: 0, shrink: 1, overflow: 'hidden', color }}
	onresize={(box: { width: number }) => marquee.measureBox(box.width)}
	onhover={followPointer}
>
	<View
		style={{ direction: 'row', shrink: 0, translate: marquee.translate() }}
		onresize={(line: { width: number }) => marquee.measureLine(line.width)}
	>
		{#each artists as artist, index (index)}
			{#if index > 0}
				<Label {size} style={{ ...oneLine, marginRight: 4 }}>,</Label>
			{/if}
			{#if artist.id}
				<Button
					style={{ shrink: 0 }}
					hover={{ color: colors.text, underline: true }}
					onpress={() => goto(itemHref('artist', artist.id))}
				>
					<Label {size} style={oneLine}>{artist.name}</Label>
				</Button>
			{:else}
				<Label {size} style={oneLine}>{artist.name}</Label>
			{/if}
		{/each}
	</View>
</View>
