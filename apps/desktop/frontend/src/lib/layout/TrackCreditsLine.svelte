<!--
  Stand-in for components/TrackCredits.svelte until the page agent converts it: the track's title,
  a link to its album when it has one, and its artists underneath.
-->
<script lang="ts">
	import { Button, View } from '@neoworks/svelte-gpui/ui';
	import type { Track } from '#lib/api/music.ts';
	import ArtistLinks from '#lib/components/ArtistLinks.svelte';
	import Label from '#lib/components/Label.svelte';
	import { itemHref } from '#lib/format.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import type { TextSize } from '#lib/theme.ts';

	let {
		track,
		titleSize = 'medium',
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
		<Button
			style={{ minWidth: 0 }}
			hover={{ underline: true }}
			onpress={() => goto(itemHref('album', albumId))}
		>
			<Label size={titleSize} weight="medium" tone="default" accented={accentTitle} truncate>
				{track.title}
			</Label>
		</Button>
	{:else}
		<Label size={titleSize} weight="medium" tone="default" accented={accentTitle} truncate>
			{track.title}
		</Label>
	{/if}
	<View style={{ color: appearance.dim, minWidth: 0 }}>
		<ArtistLinks artists={track.artists} size={artistSize} />
	</View>
</View>
