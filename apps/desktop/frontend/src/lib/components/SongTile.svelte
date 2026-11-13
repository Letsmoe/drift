<script lang="ts">
	import { View, type Shadow } from '@neoworks/svelte-gpui/ui';
	import type { Track } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import ContextArea from '#lib/components/ContextArea.svelte';
	import Cover from '#lib/components/Cover.svelte';
	import Label from '#lib/components/Label.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import TrackContextMenu from '#lib/components/TrackContextMenu.svelte';
	import { formatTrackByline, trackHref, trackKindOf } from '#lib/format.ts';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { accentRing, colors, hoverTransition, radii } from '#lib/theme.ts';

	// The old layout's h-18: its theme's spacing unit makes numbers it has no token for 3.5 px wide.
	const tileHeight = 63;

	let {
		track,
		tracks = [track],
		index = 0,
		context = 'track',
		onplay,
		onvisible
	}: {
		track: Track;
		tracks?: Track[];
		index?: number;
		context?: string;
		onplay?: () => void;
		onvisible?: () => void;
	} = $props();

	const contextMenu = new PopupMenu();
	let hovered = $state(false);

	const current = $derived(player.isCurrent(track));
	const playing = $derived(player.isPlaying(track));
	const pageHref = $derived(trackHref(track));
	const showPlayButton = $derived(hovered || playing);
	const ring = $derived(ringFor(current));

	function ringFor(isCurrent: boolean): Shadow | null {
		if (isCurrent) {
			return accentRing(appearance.accent);
		}
		return null;
	}

	function backgroundFor(isCurrent: boolean): string {
		if (isCurrent) {
			return colors.hover;
		}
		return colors.raised;
	}

	function play(): void {
		onplay?.();
		player.playOrToggle(track, tracks, index, context, trackKindOf(track));
	}
</script>

<ContextArea popup={contextMenu} onhover={(isHovered) => (hovered = isHovered)}>
	<View
		style={{
			cursor: 'pointer',
			height: tileHeight,
			direction: 'row',
			align: 'center',
			gap: 12,
			paddingRight: 8,
			radius: radii.medium,
			overflow: 'hidden',
			background: backgroundFor(current),
			shadow: ring,
			transition: hoverTransition
		}}
		hover={{ background: colors.hover }}
		onpress={() => goto(pageHref)}
	>
		<Cover
			thumbnails={track.thumbnails}
			size={tileHeight}
			style={{ width: tileHeight, height: tileHeight, shrink: 0 }}
			onload={onvisible}
		/>
		<View style={{ flex: 1, minWidth: 0 }}>
			<Label size="body" weight="semibold" accented={current} truncate>{track.title}</Label>
			<Label size="small" tone="dim" truncate>{formatTrackByline(track)}</Label>
		</View>
		{#if showPlayButton}
			<PlayOverlayButton size="sm" {playing} onpress={play} />
		{/if}
	</View>
</ContextArea>

<TrackContextMenu popup={contextMenu} {track} />
