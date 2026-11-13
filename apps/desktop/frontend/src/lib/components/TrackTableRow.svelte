<script lang="ts">
	import { Button, Icon, View } from '@neoworks/svelte-gpui/ui';
	import type { ItemKind, Track } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import ArtistLinks from '#lib/components/ArtistLinks.svelte';
	import ContextArea from '#lib/components/ContextArea.svelte';
	import Cover from '#lib/components/Cover.svelte';
	import DownloadButton from '#lib/components/DownloadButton.svelte';
	import DownloadedIndicator from '#lib/components/DownloadedIndicator.svelte';
	import IconAction from '#lib/components/IconAction.svelte';
	import Label from '#lib/components/Label.svelte';
	import Menu from '#lib/components/Menu.svelte';
	import MoreMenuButton from '#lib/components/MoreMenuButton.svelte';
	import {
		actionColumnWidth,
		rowGap,
		rowPaddingX,
		trackRowHeight
	} from '#lib/components/trackColumns.ts';
	import { formatDuration, itemHref } from '#lib/format.ts';
	import { heart, heartFill, pauseFill, playFill, speakerHighFill } from '#lib/icons.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';
	import { buildTrackMenuEntries } from '#lib/menu/trackMenuEntries.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	let {
		track,
		index,
		kind = 'song',
		showCover,
		showAlbum,
		dateAdded = undefined,
		onplay,
		onremove = undefined
	}: {
		track: Track;
		index: number;
		kind?: ItemKind;
		showCover: boolean;
		showAlbum: boolean;
		/** Set to show the "Date added" column. */
		dateAdded?: string;
		onplay: () => void;
		onremove?: () => void;
	} = $props();

	const popup = new PopupMenu();
	let pointerOver = $state(false);
	let moreMenuOpen = $state(false);

	const current = $derived(player.isCurrent(track));
	const playing = $derived(player.isPlaying(track));
	const liked = $derived(library.isLiked(track.videoId));
	const active = $derived(pointerOver || moreMenuOpen || popup.isOpen);
	const heartVisible = $derived(active || liked);
	const albumId = $derived(albumIdOf(track));

	function rowBackground(isActive: boolean): string {
		if (isActive) {
			return colors.hover;
		}
		return 'transparent';
	}

	function albumIdOf(shown: Track): string {
		if (shown.album && shown.album.id) {
			return shown.album.id;
		}
		return '';
	}

	function opacityFor(visible: boolean): number {
		if (visible) {
			return 1;
		}
		return 0;
	}

	function playPauseIcon(isPlaying: boolean): string {
		if (isPlaying) {
			return pauseFill;
		}
		return playFill;
	}

	function heartIcon(isLiked: boolean): string {
		if (isLiked) {
			return heartFill;
		}
		return heart;
	}

	function heartColor(isLiked: boolean): string {
		if (isLiked) {
			return appearance.accent;
		}
		return appearance.dim;
	}

	function heartHoverColor(isLiked: boolean): string {
		if (isLiked) {
			return appearance.accent;
		}
		return colors.text;
	}

	function playOnDoubleClick(event: { clicks: number }): void {
		if (event.clicks >= 2) {
			onplay();
		}
	}

	function entries() {
		return buildTrackMenuEntries(track, { kind, onRemoveFromPlaylist: onremove });
	}
</script>

<ContextArea
	{popup}
	onhover={(isHovered) => (pointerOver = isHovered)}
	onpress={playOnDoubleClick}
	style={{
		height: trackRowHeight,
		direction: 'row',
		align: 'center',
		gap: rowGap,
		paddingX: rowPaddingX,
		radius: radii.medium,
		background: rowBackground(active),
		transition: hoverTransition
	}}
>
	<View style={{ width: 16, shrink: 0, align: 'end', justify: 'center' }}>
		{#if active}
			<Button style={{ color: colors.text }} onpress={onplay}>
				<Icon svg={playPauseIcon(playing)} size={16} />
			</Button>
		{:else if playing}
			<Icon svg={speakerHighFill} size={16} color={appearance.accent} />
		{:else}
			<Label size="small" tone="dim" accented={current}>{index + 1}</Label>
		{/if}
	</View>

	<View style={{ flex: 6, minWidth: 0, direction: 'row', align: 'center', gap: 12 }}>
		{#if showCover}
			<Cover
				thumbnails={track.thumbnails}
				size={40}
				style={{ size: 40, shrink: 0, radius: radii.small }}
			/>
		{/if}
		<View style={{ flex: 1, minWidth: 0 }}>
			<View style={{ direction: 'row', align: 'center', gap: 6, minWidth: 0 }}>
				<Label size="body" accented={current} truncate style={{ flex: 1, minWidth: 0 }}>
					{track.title}
				</Label>
				<DownloadedIndicator videoId={track.videoId} />
			</View>
			<View style={{ direction: 'row', align: 'center', gap: 5, minWidth: 0 }}>
				<ArtistLinks artists={track.artists} color={appearance.dim} />
				{#if track.playCount}
					<Label size="small" tone="dim" truncate style={{ flex: 1, minWidth: 0 }}>
						• {track.playCount}
					</Label>
				{/if}
			</View>
		</View>
	</View>

	{#if showAlbum}
		<View style={{ flex: 4, minWidth: 0, direction: 'row' }}>
			{#if albumId !== '' && track.album}
				<View
					style={{ flex: 1, minWidth: 0, cursor: 'pointer', color: appearance.dim }}
					hover={{ color: colors.text, underline: true }}
					onpress={() => goto(itemHref('album', albumId))}
				>
					<Label size="small" truncate>{track.album.name}</Label>
				</View>
			{/if}
		</View>
	{/if}

	{#if dateAdded !== undefined}
		<View style={{ flex: 3, minWidth: 0 }}>
			<Label size="small" tone="dim" truncate>{dateAdded}</Label>
		</View>
	{/if}

	<View
		style={{
			width: actionColumnWidth,
			shrink: 0,
			direction: 'row',
			align: 'center',
			justify: 'end',
			gap: 12
		}}
	>
		<DownloadButton {track} {kind} rowHovered={active} />
		<IconAction
			icon={heartIcon(liked)}
			iconSize={18}
			color={heartColor(liked)}
			hoverColor={heartHoverColor(liked)}
			style={{ opacity: opacityFor(heartVisible) }}
			onpress={() => library.toggleLike(track, kind)}
		/>
		<Label size="small" tone="dim" style={{ width: 40, textAlign: 'right', shrink: 0 }}>
			{formatDuration(track.durationSeconds)}
		</Label>
		<MoreMenuButton
			bind:open={moreMenuOpen}
			buildEntries={entries}
			size={22}
			boxSize={22}
			style={{ opacity: opacityFor(active) }}
		/>
	</View>
</ContextArea>

{#if popup.position}
	<Menu
		entries={entries()}
		x={popup.position.x}
		y={popup.position.y}
		onclose={() => popup.close()}
	/>
{/if}
