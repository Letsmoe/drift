<!-- Seek bar, transport buttons, like and add to playlist of the full-screen player. -->
<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { LoadingSpinner } from '@neoworks/svelte-gpui-ui';
	import type { ItemKind } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import AddToPlaylistButton from '#lib/components/AddToPlaylistButton.svelte';
	import IconAction from '#lib/components/IconAction.svelte';
	import Label from '#lib/components/Label.svelte';
	import LikeButton from '#lib/components/LikeButton.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import RangeSlider from '#lib/components/RangeSlider.svelte';
	import ToggleIconButton from '#lib/components/ToggleIconButton.svelte';
	import { formatDuration } from '#lib/format.ts';
	import {
		repeat,
		repeatBold,
		repeatOnce,
		repeatOnceBold,
		skipBackFill,
		skipForwardFill
	} from '#lib/layout/chromeIcons.ts';
	import { shuffleActiveIcon, shuffleIcon, shuffleLabel } from '#lib/layout/shuffleIcons.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { colors } from '#lib/theme.ts';

	const kind = $derived(currentKind());
	const repeatIcon = $derived(pickRepeatIcon());
	const repeatActiveIcon = $derived(pickRepeatActiveIcon());

	function currentKind(): ItemKind {
		if (!player.entry) {
			return 'song';
		}
		return player.entry.kind as ItemKind;
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
</script>

<View style={{ width: 'full', gap: 12 }}>
	<View style={{ direction: 'row', align: 'center', gap: 12 }}>
		<View style={{ width: 40, align: 'end' }}>
			<Label size="tiny" tone="muted" style={{ fontFamily: 'mono' }}>
				{formatDuration(Math.floor(player.currentTime)) || '0:00'}
			</Label>
		</View>
		<RangeSlider
			value={player.currentTime}
			max={player.duration || 0}
			disabled={!player.track}
			oncommit={(seconds) => player.seek(seconds)}
		/>
		<View style={{ width: 40 }}>
			<Label size="tiny" tone="muted" style={{ fontFamily: 'mono' }}>
				{formatDuration(Math.floor(player.duration))}
			</Label>
		</View>
	</View>

	{#if player.errorMessage}
		<Label size="tiny" tone="danger" style={{ textAlign: 'center' }}>{player.errorMessage}</Label>
	{/if}

	<View style={{ direction: 'row', align: 'center', justify: 'center', gap: 12 }}>
		{#if player.track}
			<AddToPlaylistButton track={player.track} large />
		{/if}
		<ToggleIconButton
			icon={shuffleIcon(player.queue)}
			activeIcon={shuffleActiveIcon(player.queue)}
			label={shuffleLabel(player.queue)}
			active={player.queue.shuffle}
			onpress={() => player.cycleShuffle()}
		/>
		<IconAction
			icon={skipBackFill}
			iconSize={20}
			boxSize={40}
			hoverBackground={colors.hover}
			onpress={() => player.previous()}
		/>
		{#if player.buffering && player.track && !player.errorMessage}
			<View style={{ size: 56, align: 'center', justify: 'center' }}>
				<LoadingSpinner size={24} color={appearance.accent} />
			</View>
		{:else}
			<PlayOverlayButton size="lg" playing={!player.paused} onpress={() => player.toggle()} />
		{/if}
		<IconAction
			icon={skipForwardFill}
			iconSize={20}
			boxSize={40}
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
		{#if player.track}
			<LikeButton track={player.track} {kind} size={24} />
		{/if}
	</View>
</View>
