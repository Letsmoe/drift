<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import type { Item, ItemKind } from '#lib/api/music.ts';
	import ContextArea from '#lib/components/ContextArea.svelte';
	import Cover from '#lib/components/Cover.svelte';
	import Label from '#lib/components/Label.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import TrackContextMenu from '#lib/components/TrackContextMenu.svelte';
	import { itemHref, trackHref } from '#lib/format.ts';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	// Electron's w-56: the old theme's spacing unit makes numbers it has no token for 3.5 px wide.
	const cardWidth = 196;
	const coverSide = cardWidth - 24;
	// Thumbnail to fetch, sharp enough for 2x screens.
	const coverSize = 200;

	let { item, context = 'home' }: { item: Item; context?: string } = $props();

	const contextMenu = new PopupMenu();
	let hovered = $state(false);

	const playing = $derived(isTrackPlaying());
	// Songs and episodes open their album or show; the play button plays them in place.
	const pageHref = $derived(hrefOf(item));
	const subtitle = $derived(subtitleOf(item));
	const coverRadius = $derived(coverRadiusOf(item));
	const showPlayButton = $derived(item.track !== undefined && (hovered || playing));

	// Song cards already carry "Artist • 6.2M plays" in YouTube's subtitle; episodes don't.
	function subtitleOf(shown: Item): string {
		const playCount = shown.track?.playCount;
		if (!playCount || shown.subtitle.includes(playCount)) {
			return shown.subtitle;
		}
		return [shown.subtitle, playCount].filter((part) => part !== '').join(' • ');
	}

	function hrefOf(shown: Item): string | undefined {
		if (shown.track) {
			return trackHref(shown.track);
		}
		return itemHref(shown.kind, shown.id);
	}

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

	function openPage(): void {
		if (pageHref) {
			goto(pageHref);
		}
	}

	function playTrack(): void {
		if (item.track) {
			player.playOrToggle(item.track, [item.track], 0, context, item.kind as ItemKind);
		}
	}
</script>

<ContextArea
	popup={item.track ? contextMenu : undefined}
	onhover={(isHovered) => (hovered = isHovered)}
	style={{ width: cardWidth, shrink: 0 }}
>
	<View
		style={{
			cursor: 'pointer',
			align: 'stretch',
			gap: 8,
			padding: 12,
			radius: radii.large,
			transition: hoverTransition
		}}
		hover={{ background: colors.hover }}
		onpress={openPage}
	>
		<View style={{ position: 'relative' }}>
			<Cover
				thumbnails={item.thumbnails}
				size={coverSize}
				style={{
					radius: coverRadius,
					shadow: { y: 4, blur: 12, color: 'rgba(0, 0, 0, 0.45)' }
				}}
			/>
			{#if showPlayButton}
				<PlayOverlayButton
					{playing}
					onpress={playTrack}
					style={{ position: 'absolute', right: 8, bottom: 8 }}
				/>
			{/if}
		</View>
		<View style={{ minWidth: 0 }}>
			<Label size="body" weight="medium" truncate>{item.title}</Label>
			<Label size="small" tone="dim" lines={2}>{subtitle}</Label>
		</View>
	</View>
</ContextArea>

{#if item.track}
	<TrackContextMenu popup={contextMenu} track={item.track} kind={item.kind as ItemKind} />
{/if}
