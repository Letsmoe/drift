<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import type { Track } from '#lib/api/music.ts';
	import ContextArea from '#lib/components/ContextArea.svelte';
	import Cover from '#lib/components/Cover.svelte';
	import Label from '#lib/components/Label.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import TrackContextMenu from '#lib/components/TrackContextMenu.svelte';
	import { formatTrackByline, trackHref } from '#lib/format.ts';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	const coverSize = 184;

	let { track, onplay, onvisible }: { track: Track; onplay: () => void; onvisible?: () => void } =
		$props();

	const contextMenu = new PopupMenu();
	let hovered = $state(false);

	const playing = $derived(player.isPlaying(track));
	const current = $derived(player.isCurrent(track));
	const showPlayButton = $derived(hovered || playing);
</script>

<ContextArea
	popup={contextMenu}
	onhover={(isHovered) => (hovered = isHovered)}
	style={{ width: 182, shrink: 0 }}
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
		hover={{ background: colors.raised }}
		onpress={() => goto(trackHref(track))}
	>
		<View style={{ position: 'relative' }}>
			<Cover
				thumbnails={track.thumbnails}
				size={coverSize}
				style={{ radius: radii.medium }}
				onload={onvisible}
			/>
			{#if showPlayButton}
				<PlayOverlayButton
					{playing}
					onpress={onplay}
					style={{ position: 'absolute', right: 8, bottom: 8 }}
				/>
			{/if}
		</View>
		<View style={{ minWidth: 0 }}>
			<Label size="small" weight="semibold" accented={current} truncate>{track.title}</Label>
			<Label size="tiny" tone="dim" truncate>{formatTrackByline(track)}</Label>
		</View>
	</View>
</ContextArea>

<TrackContextMenu popup={contextMenu} {track} />
