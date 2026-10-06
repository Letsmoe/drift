<script lang="ts">
	import { Button, LoadingSpinner } from '@neoworks-dev/ui';
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import QueueIcon from 'phosphor-svelte/lib/QueueIcon';
	import RepeatIcon from 'phosphor-svelte/lib/RepeatIcon';
	import RepeatOnceIcon from 'phosphor-svelte/lib/RepeatOnceIcon';
	import ShuffleIcon from 'phosphor-svelte/lib/ShuffleIcon';
	import SkipBackIcon from 'phosphor-svelte/lib/SkipBackIcon';
	import SkipForwardIcon from 'phosphor-svelte/lib/SkipForwardIcon';
	import SpeakerHighIcon from 'phosphor-svelte/lib/SpeakerHighIcon';
	import SpeakerLowIcon from 'phosphor-svelte/lib/SpeakerLowIcon';
	import SpeakerXIcon from 'phosphor-svelte/lib/SpeakerXIcon';
	import AddToPlaylistButton from '#lib/components/AddToPlaylistButton.svelte';
	import Cover from '#lib/components/Cover.svelte';
	import DownloadButton from '#lib/components/DownloadButton.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import RangeSlider from '#lib/components/RangeSlider.svelte';
	import TrackCredits from '#lib/components/TrackCredits.svelte';
	import ToggleIconButton from '#lib/components/ToggleIconButton.svelte';
	import { formatDuration } from '#lib/format.ts';
	import { panels } from '#lib/layout/panels.svelte.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import type { ItemKind } from '#lib/api/music.ts';

	const VolumeIcon = $derived(pickVolumeIcon());
	const liked = $derived(player.track !== null && library.isLiked(player.track.videoId));
	const likeLabel = $derived(buildLikeLabel());
	const likeWeight = $derived(buildLikeWeight());
	const RepeatButtonIcon = $derived(pickRepeatIcon());

	function pickRepeatIcon() {
		if (player.queue.repeat === 'one') {
			return RepeatOnceIcon;
		}
		return RepeatIcon;
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

	function shuffleLabel(): string {
		if (player.queue.shuffle) {
			return 'Shuffle: on';
		}
		return 'Shuffle: off';
	}

	function currentKind(): ItemKind {
		if (!player.entry) {
			return 'song';
		}
		return player.entry.kind as ItemKind;
	}

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

	function buildLikeLabel(): string {
		if (liked) {
			return 'Remove from Liked Songs';
		}
		return 'Save to Liked Songs';
	}

	function buildLikeWeight(): 'fill' | 'regular' {
		if (liked) {
			return 'fill';
		}
		return 'regular';
	}

	function toggleLikeOfCurrent(): void {
		if (!player.track || !player.entry) {
			return;
		}
		library.toggleLike(player.track, player.entry.kind as ItemKind);
	}

	function isTypingTarget(target: EventTarget | null): boolean {
		if (!(target instanceof HTMLElement)) {
			return false;
		}
		return target.closest('input, textarea, select, button, a, [contenteditable]') !== null;
	}

	function handleKeydown(event: KeyboardEvent): void {
		if (event.code !== 'Space' || isTypingTarget(event.target)) {
			return;
		}
		event.preventDefault();
		player.toggle();
	}

	// Playback position at the last `waiting` or `timeupdate`.
	let lastPosition = 0;

	function startBuffering(event: Event): void {
		const audio = event.currentTarget as HTMLAudioElement;
		lastPosition = audio.currentTime;
		player.buffering = true;
	}

	// After a seek WebKitGTK keeps playing but neither fires `playing` nor raises readyState
	// past HAVE_CURRENT_DATA, so buffering ends once the position moves on with data loaded.
	function endBufferingWhenAdvancing(event: Event): void {
		const audio = event.currentTarget as HTMLAudioElement;
		const advanced = audio.currentTime !== lastPosition;
		lastPosition = audio.currentTime;
		if (advanced && !audio.paused && audio.readyState >= HTMLMediaElement.HAVE_CURRENT_DATA) {
			player.buffering = false;
		}
	}

	function shownVolume(): number {
		if (player.muted) {
			return 0;
		}
		return player.volume;
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<audio
	src={player.source}
	autoplay
	bind:paused={player.paused}
	bind:currentTime={player.currentTime}
	bind:duration={player.duration}
	bind:volume={player.volume}
	bind:muted={player.muted}
	onwaiting={startBuffering}
	onplaying={() => (player.buffering = false)}
	ontimeupdate={endBufferingWhenAdvancing}
	onended={() => player.trackEnded()}
	onerror={() => player.fail('This song could not be played.')}
></audio>

<footer class="grid h-18 shrink-0 grid-cols-[1fr_minmax(0,2fr)_1fr] items-center gap-4 px-2">
	<div class="flex min-w-0 items-center gap-3">
		{#if player.track}
			<button
				type="button"
				aria-label="Open full-screen player"
				onclick={() => panels.toggleFullPlayer()}
				class="shrink-0 cursor-pointer rounded-md"
			>
				<Cover thumbnails={player.track.thumbnails} size={56} class="size-14 rounded-md" />
			</button>
			<div class="min-w-0">
				<TrackCredits track={player.track} />
			</div>
			<div class="flex shrink-0 items-center">
				<button
					type="button"
					aria-label={likeLabel}
					title={likeLabel}
					aria-pressed={liked}
					onclick={toggleLikeOfCurrent}
					class={[
						'duration-fast flex size-8 items-center justify-center rounded-full transition-colors hover:bg-hover',
						liked && 'text-action',
						!liked && 'text-muted hover:text-default'
					]}
				>
					<HeartIcon size={20} weight={likeWeight} />
				</button>
				<AddToPlaylistButton track={player.track} />
				<span class="flex size-8 items-center justify-center text-muted">
					<DownloadButton
						track={player.track}
						kind={currentKind()}
						size={20}
						revealOnHover={false}
					/>
				</span>
			</div>
		{/if}
	</div>

	<div class="flex flex-col items-center gap-1">
		<div class="flex items-center gap-2">
			<ToggleIconButton
				icon={ShuffleIcon}
				label={shuffleLabel()}
				active={player.queue.shuffle}
				onclick={() => player.toggleShuffle()}
			/>
			<Button
				round
				size="sm"
				variant="ghost"
				icon={SkipBackIcon}
				onclick={() => player.previous()}
			/>
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
			<Button
				round
				size="sm"
				variant="ghost"
				icon={SkipForwardIcon}
				onclick={() => player.next()}
			/>
			<ToggleIconButton
				icon={RepeatButtonIcon}
				label={repeatLabel()}
				active={player.queue.repeat !== 'off'}
				onclick={() => player.cycleRepeat()}
			/>
		</div>

		{#if player.errorMessage}
			<p class="h-4 text-xs text-red">{player.errorMessage}</p>
		{:else}
			<div class="flex w-full items-center gap-2">
				<span class="w-10 text-right text-xs text-dim tabular-nums">
					{formatDuration(Math.floor(player.currentTime))}
				</span>
				<RangeSlider
					label="Seek"
					value={player.currentTime}
					max={player.duration || 0}
					disabled={!player.track}
					oninput={(seconds) => player.seek(seconds)}
				/>
				<span class="w-10 text-xs text-dim tabular-nums">
					{formatDuration(Math.floor(player.duration))}
				</span>
			</div>
		{/if}
	</div>

	<div class="flex items-center justify-end gap-1 pr-2">
		<ToggleIconButton
			icon={QueueIcon}
			label="Queue"
			active={panels.queueOpen}
			onclick={() => panels.toggleQueue()}
		/>
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
