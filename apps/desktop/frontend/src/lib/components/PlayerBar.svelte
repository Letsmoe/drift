<script lang="ts">
	import { Button, LoadingSpinner } from '@neoworks-dev/ui';
	import RepeatIcon from 'phosphor-svelte/lib/RepeatIcon';
	import ShuffleIcon from 'phosphor-svelte/lib/ShuffleIcon';
	import SkipBackIcon from 'phosphor-svelte/lib/SkipBackIcon';
	import SkipForwardIcon from 'phosphor-svelte/lib/SkipForwardIcon';
	import SpeakerHighIcon from 'phosphor-svelte/lib/SpeakerHighIcon';
	import SpeakerLowIcon from 'phosphor-svelte/lib/SpeakerLowIcon';
	import SpeakerXIcon from 'phosphor-svelte/lib/SpeakerXIcon';
	import Cover from '#lib/components/Cover.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import RangeSlider from '#lib/components/RangeSlider.svelte';
	import { formatArtists, formatDuration } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';

	const VolumeIcon = $derived(pickVolumeIcon());

	function pickVolumeIcon() {
		if (player.muted || player.volume === 0) {
			return SpeakerXIcon;
		}
		if (player.volume < 0.5) {
			return SpeakerLowIcon;
		}
		return SpeakerHighIcon;
	}

	function playButtonLabel(): string {
		if (player.paused) {
			return 'Play';
		}
		return 'Pause';
	}

	function shownVolume(): number {
		if (player.muted) {
			return 0;
		}
		return player.volume;
	}
</script>

<audio
	src={player.source}
	autoplay
	bind:paused={player.paused}
	bind:currentTime={player.currentTime}
	bind:duration={player.duration}
	bind:volume={player.volume}
	bind:muted={player.muted}
	onwaiting={() => (player.buffering = true)}
	onplaying={() => (player.buffering = false)}
	onerror={() => player.fail('This song could not be played.')}
></audio>

<footer class="grid h-18 shrink-0 grid-cols-[1fr_minmax(0,2fr)_1fr] items-center gap-4 px-2">
	<div class="flex min-w-0 items-center gap-3">
		{#if player.track}
			<Cover thumbnails={player.track.thumbnails} size={56} class="size-14 shrink-0 rounded-md" />
			<div class="min-w-0">
				<p class="truncate text-base font-medium">{player.track.title}</p>
				<p class="truncate text-sm text-dim">{formatArtists(player.track.artists)}</p>
			</div>
		{/if}
	</div>

	<div class="flex flex-col items-center gap-1">
		<div class="flex items-center gap-2">
			<Button round size="sm" variant="ghost" icon={ShuffleIcon} disabled />
			<Button round size="sm" variant="ghost" icon={SkipBackIcon} disabled />
			{#if player.buffering && player.track && !player.errorMessage}
				<div class="flex size-8 items-center justify-center"><LoadingSpinner size={18} /></div>
			{:else}
				<PlayOverlayButton
					size="sm"
					playing={!player.paused}
					label={playButtonLabel()}
					onclick={() => player.toggle()}
				/>
			{/if}
			<Button round size="sm" variant="ghost" icon={SkipForwardIcon} disabled />
			<Button round size="sm" variant="ghost" icon={RepeatIcon} disabled />
		</div>

		{#if player.errorMessage}
			<p class="h-4 text-xs text-red">{player.errorMessage}</p>
		{:else}
			<div class="flex w-full items-center gap-2">
				<span class="w-10 text-right font-mono text-xs text-dim tabular-nums">
					{formatDuration(Math.floor(player.currentTime))}
				</span>
				<RangeSlider
					label="Seek"
					value={player.currentTime}
					max={player.duration || 0}
					disabled={!player.track}
					oninput={(seconds) => player.seek(seconds)}
				/>
				<span class="w-10 font-mono text-xs text-dim tabular-nums">
					{formatDuration(Math.floor(player.duration))}
				</span>
			</div>
		{/if}
	</div>

	<div class="flex items-center justify-end gap-1 pr-2">
		<Button round size="sm" variant="ghost" icon={VolumeIcon} onclick={() => player.toggleMute()} />
		<div class="w-24">
			<RangeSlider
				label="Volume"
				value={shownVolume()}
				max={1}
				step={0.01}
				oninput={(volume) => player.setVolume(volume)}
			/>
		</div>
	</div>
</footer>
