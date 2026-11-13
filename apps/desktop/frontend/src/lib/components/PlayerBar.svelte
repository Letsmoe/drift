<script lang="ts">
	import { Button, Icon, View } from '@neoworks/svelte-gpui/ui';
	import { LoadingSpinner } from '@neoworks/svelte-gpui-ui';
	import type { ItemKind } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import AddToPlaylistButton from '#lib/components/AddToPlaylistButton.svelte';
	import Cover from '#lib/components/Cover.svelte';
	import DownloadButton from '#lib/components/DownloadButton.svelte';
	import IconAction from '#lib/components/IconAction.svelte';
	import Label from '#lib/components/Label.svelte';
	import LikeButton from '#lib/components/LikeButton.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import RangeSlider from '#lib/components/RangeSlider.svelte';
	import ToggleIconButton from '#lib/components/ToggleIconButton.svelte';
	import { formatDuration } from '#lib/format.ts';
	import {
		queue,
		repeat,
		repeatBold,
		repeatOnce,
		repeatOnceBold,
		skipBackFill,
		skipForwardFill,
		speakerHigh,
		speakerLow,
		speakerX
	} from '#lib/layout/chromeIcons.ts';
	import { panels } from '#lib/layout/panels.svelte.ts';
	import { shuffleActiveIcon, shuffleIcon, shuffleLabel } from '#lib/layout/shuffleIcons.ts';
	import TrackCreditsLine from '#lib/layout/TrackCreditsLine.svelte';
	import { connectAudio } from '#lib/player/audioBinding.svelte.ts';
	import { Crossfade } from '#lib/player/crossfade.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { colors, radii } from '#lib/theme.ts';

	const crossfade = new Crossfade(player);
	connectAudio(player, crossfade);

	const barHeight = 72;
	const coverSize = 56;

	const volumeIcon = $derived(pickVolumeIcon());
	const repeatIcon = $derived(pickRepeatIcon());
	const repeatActiveIcon = $derived(pickRepeatActiveIcon());
	const shownVolume = $derived(pickShownVolume());
	const kind = $derived(currentKind());

	function pickVolumeIcon(): string {
		if (player.muted || player.volume === 0) {
			return speakerX;
		}
		if (player.volume < 0.5) {
			return speakerLow;
		}
		return speakerHigh;
	}

	function pickRepeatIcon(): string {
		if (player.queue.repeat === 'one') {
			return repeatOnce;
		}
		return repeat;
	}

	function pickRepeatActiveIcon(): string {
		if (player.queue.repeat === 'one') {
			return repeatOnceBold;
		}
		return repeatBold;
	}

	function repeatLabel(): string {
		if (player.queue.repeat === 'one') {
			return 'Repeat: one song';
		}
		if (player.queue.repeat === 'off') {
			return 'Repeat: off';
		}
		return 'Repeat: all';
	}

	function pickShownVolume(): number {
		if (player.muted) {
			return 0;
		}
		return player.volume;
	}

	function currentKind(): ItemKind {
		if (!player.entry) {
			return 'song';
		}
		return player.entry.kind as ItemKind;
	}
</script>

<View
	style={{
		direction: 'row',
		align: 'center',
		height: barHeight,
		shrink: 0,
		gap: 16,
		paddingX: 8
	}}
>
	<View style={{ flex: 1, minWidth: 0, direction: 'row', align: 'center', gap: 12 }}>
		{#if player.track}
			<Button
				style={{ shrink: 0, radius: radii.medium, overflow: 'hidden' }}
				onpress={() => panels.toggleFullPlayer()}
			>
				<Cover
					thumbnails={player.track.thumbnails}
					size={coverSize}
					style={{ size: coverSize, radius: radii.medium }}
				/>
			</Button>
			<View style={{ flex: 1, minWidth: 0 }}>
				<TrackCreditsLine track={player.track} />
			</View>
			<View style={{ direction: 'row', align: 'center', shrink: 0 }}>
				<LikeButton track={player.track} {kind} />
				<AddToPlaylistButton track={player.track} />
				<View style={{ size: 32, align: 'center', justify: 'center' }}>
					<DownloadButton track={player.track} {kind} size={20} revealOnHover={false} />
				</View>
			</View>
		{/if}
	</View>

	<View style={{ flex: 2, minWidth: 0, align: 'center', gap: 4 }}>
		<View
			style={{
				direction: 'row',
				align: 'center',
				gap: 8,
				opacity: player.track ? 1 : 0.4
			}}
		>
			<ToggleIconButton
				icon={shuffleIcon(player.queue)}
				activeIcon={shuffleActiveIcon(player.queue)}
				label={shuffleLabel(player.queue)}
				active={player.queue.shuffle}
				onpress={() => player.cycleShuffle()}
			/>
			<IconAction
				icon={skipBackFill}
				iconSize={16}
				boxSize={32}
				hoverBackground={colors.hover}
				onpress={() => player.previous()}
			/>
			{#if player.buffering && player.track && !player.errorMessage}
				<View style={{ size: 32, align: 'center', justify: 'center' }}>
					<LoadingSpinner size={18} color={appearance.accent} />
				</View>
			{:else}
				<PlayOverlayButton size="sm" playing={!player.paused} onpress={() => player.toggle()} />
			{/if}
			<IconAction
				icon={skipForwardFill}
				iconSize={16}
				boxSize={32}
				hoverBackground={colors.hover}
				onpress={() => player.next()}
			/>
			<ToggleIconButton
				icon={repeatIcon}
				activeIcon={repeatActiveIcon}
				label={repeatLabel()}
				active={player.queue.repeat !== 'off'}
				onpress={() => player.cycleRepeat()}
			/>
		</View>

		{#if player.errorMessage}
			<View style={{ height: 16 }}>
				<Label size="tiny" tone="danger">{player.errorMessage}</Label>
			</View>
		{:else if !player.track}
			<View style={{ height: 16 }}>
				<Label size="tiny" tone="dim">Pick something to play</Label>
			</View>
		{:else}
			<View style={{ direction: 'row', align: 'center', gap: 8, width: 'full' }}>
				<View style={{ width: 40, align: 'end' }}>
					<Label size="tiny" tone="dim"
						>{formatDuration(Math.floor(player.currentTime)) || '0:00'}</Label
					>
				</View>
				<RangeSlider
					value={player.currentTime}
					max={player.duration || 0}
					oncommit={(seconds) => player.seek(seconds)}
				/>
				<View style={{ width: 40 }}>
					<Label size="tiny" tone="dim">{formatDuration(Math.floor(player.duration))}</Label>
				</View>
			</View>
		{/if}
	</View>

	<View
		style={{
			flex: 1,
			minWidth: 0,
			direction: 'row',
			align: 'center',
			justify: 'end',
			gap: 4,
			paddingRight: 8
		}}
	>
		<ToggleIconButton
			icon={queue}
			label="Queue"
			active={panels.queueOpen}
			onpress={() => panels.toggleQueue()}
		/>
		<IconAction
			icon={volumeIcon}
			iconSize={20}
			boxSize={32}
			hoverBackground={colors.hover}
			onpress={() => player.toggleMute()}
		/>
		<View style={{ width: 96, direction: 'row' }}>
			<RangeSlider
				value={shownVolume}
				max={1}
				step={0.01}
				oninput={(volume) => player.setVolume(volume)}
			/>
		</View>
	</View>
</View>
