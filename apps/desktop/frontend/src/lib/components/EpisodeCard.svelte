<script lang="ts">
	import { View, type Shadow } from '@neoworks/svelte-gpui/ui';
	import type { Item, Track } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import ContextArea from '#lib/components/ContextArea.svelte';
	import Cover from '#lib/components/Cover.svelte';
	import Label from '#lib/components/Label.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import TrackContextMenu from '#lib/components/TrackContextMenu.svelte';
	import { formatDuration, playCountOf, trackHref } from '#lib/format.ts';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { accentRing, colors, hoverTransition, radii } from '#lib/theme.ts';

	const coverWidth = 144;

	let {
		item,
		track,
		tracks = [track],
		index = 0,
		context = 'home'
	}: { item: Item; track: Track; tracks?: Track[]; index?: number; context?: string } = $props();

	const contextMenu = new PopupMenu();
	let hovered = $state(false);

	const playing = $derived(player.isPlaying(track));
	const current = $derived(player.isCurrent(track));
	const showPlayButton = $derived(hovered || playing);
	const ring = $derived(ringFor(current));
	const details = $derived(
		[publishedOf(item), formatDuration(track.durationSeconds), playCountOf(track)]
			.filter((part) => part !== '')
			.join(' • ')
	);

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

	function publishedOf(shown: Item): string {
		if (shown.published) {
			return shown.published;
		}
		return '';
	}

	function play(): void {
		player.playOrToggle(track, tracks, index, context, 'episode');
	}
</script>

<ContextArea popup={contextMenu} onhover={(isHovered) => (hovered = isHovered)} style={{ flex: 1 }}>
	<View
		style={{
			cursor: 'pointer',
			flex: 1,
			direction: 'row',
			align: 'stretch',
			gap: 12,
			padding: 8,
			radius: radii.large,
			background: backgroundFor(current),
			shadow: ring,
			transition: hoverTransition
		}}
		hover={{ background: colors.hover }}
		onpress={() => goto(trackHref(track))}
	>
		<View style={{ width: coverWidth, shrink: 0, alignSelf: 'start', position: 'relative' }}>
			<Cover
				thumbnails={track.thumbnails}
				size={coverWidth}
				aspect="video"
				style={{ radius: radii.medium }}
			/>
			{#if showPlayButton}
				<PlayOverlayButton
					size="sm"
					{playing}
					onpress={play}
					style={{ position: 'absolute', right: 6, bottom: 6 }}
				/>
			{/if}
		</View>
		<View style={{ flex: 1, minWidth: 0, gap: 2 }}>
			<Label size="small" weight="semibold" accented={current} lines={2}>{track.title}</Label>
			<Label size="tiny" tone="muted" truncate>{item.subtitle}</Label>
			{#if item.description}
				<Label size="tiny" tone="dim" lines={2}>{item.description}</Label>
			{/if}
			<Label size="tiny" tone="faint" truncate style={{ paddingTop: 4 }}>{details}</Label>
		</View>
	</View>
</ContextArea>

<TrackContextMenu popup={contextMenu} {track} kind="episode" />
