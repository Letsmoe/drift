<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import type { Item, ItemKind } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import Label from '#lib/components/Label.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { itemHref } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	const coverSide = 96;

	let { item, context }: { item: Item; context: string } = $props();

	let hovered = $state(false);

	const playing = $derived(isTrackPlaying());
	const showPlayButton = $derived(item.track !== undefined && (hovered || playing));
	const coverRadius = $derived(coverRadiusOf(item));

	function coverRadiusOf(shown: Item): number {
		if (shown.kind === 'artist') {
			return coverSide / 2;
		}
		return radii.medium;
	}

	function isTrackPlaying(): boolean {
		if (!item.track) {
			return false;
		}
		return player.isPlaying(item.track);
	}

	// Songs play in place; every other kind opens its page.
	function openPage(): void {
		if (item.track) {
			return;
		}
		goto(itemHref(item.kind, item.id));
	}

	function playTrack(): void {
		if (item.track) {
			player.playOrToggle(item.track, [item.track], 0, context, item.kind as ItemKind);
		}
	}
</script>

<View
	style={{
		cursor: 'pointer',
		align: 'stretch',
		gap: 20,
		padding: 20,
		position: 'relative',
		radius: radii.large,
		background: colors.raised,
		transition: hoverTransition
	}}
	hover={{ background: colors.hover }}
	onhover={(isHovered) => (hovered = isHovered)}
	onpress={openPage}
>
	<Cover
		thumbnails={item.thumbnails}
		size={coverSide}
		style={{
			size: coverSide,
			radius: coverRadius,
			shadow: { y: 4, blur: 12, color: 'rgba(0, 0, 0, 0.45)' }
		}}
	/>
	<View style={{ minWidth: 0 }}>
		<Label size="title" weight="bold" truncate>{item.title}</Label>
		<Label size="small" tone="dim" truncate>{item.subtitle}</Label>
	</View>
	{#if showPlayButton}
		<PlayOverlayButton
			{playing}
			onpress={playTrack}
			style={{ position: 'absolute', right: 20, bottom: 20 }}
		/>
	{/if}
</View>
