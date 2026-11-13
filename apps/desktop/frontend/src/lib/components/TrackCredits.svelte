<!--
  A track's title, linked to its album when it has one, and its artists underneath. The title and the artist
  links are separate presses, so the title is a view that listens for presses and not a paragraph link.
-->
<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import type { Track } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import ArtistLinks from '#lib/components/ArtistLinks.svelte';
	import Label from '#lib/components/Label.svelte';
	import { itemHref } from '#lib/format.ts';
	import { goto } from '#lib/router.svelte.ts';
	import type { TextSize } from '#lib/theme.ts';

	let {
		track,
		titleSize = 'body',
		artistSize = 'small',
		accentTitle = false
	}: {
		track: Track;
		titleSize?: TextSize;
		artistSize?: TextSize;
		/** Draws the title in the accent colour, as the current queue entry does. */
		accentTitle?: boolean;
	} = $props();

	const albumId = $derived(track.album?.id);
</script>

<View style={{ minWidth: 0 }}>
	{#if albumId}
		<View
			style={{ alignSelf: 'start', maxWidth: '100%', cursor: 'pointer' }}
			hover={{ underline: true }}
			onpress={() => goto(itemHref('album', albumId))}
		>
			<Label size={titleSize} weight="medium" accented={accentTitle} truncate>{track.title}</Label>
		</View>
	{:else}
		<Label size={titleSize} weight="medium" accented={accentTitle} truncate>{track.title}</Label>
	{/if}
	<ArtistLinks artists={track.artists} size={artistSize} color={appearance.dim} />
</View>
