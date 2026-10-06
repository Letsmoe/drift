<script lang="ts">
	import { Button, LoadingSpinner } from '@neoworks-dev/ui';
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import RepeatIcon from 'phosphor-svelte/lib/RepeatIcon';
	import RepeatOnceIcon from 'phosphor-svelte/lib/RepeatOnceIcon';
	import SkipBackIcon from 'phosphor-svelte/lib/SkipBackIcon';
	import SkipForwardIcon from 'phosphor-svelte/lib/SkipForwardIcon';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import RangeSlider from '#lib/components/RangeSlider.svelte';
	import { formatDuration } from '#lib/format.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { shuffleIcon, shuffleLabel } from '#lib/player/shuffleMode.ts';
	import type { ItemKind } from '#lib/api/music.ts';

	const liked = $derived(player.track !== null && library.isLiked(player.track.videoId));
	const RepeatButtonIcon = $derived(pickRepeatIcon());

	function pickRepeatIcon() {
		if (player.queue.repeat === 'one') {
			return RepeatOnceIcon;
		}
		return RepeatIcon;
	}

	function playButtonLabel(): string {
		if (player.paused) {
			return 'Play';
		}
		return 'Pause';
	}

	function likeLabel(): string {
		if (liked) {
			return 'Remove from Liked Songs';
		}
		return 'Save to Liked Songs';
	}

	function likeWeight(): 'fill' | 'regular' {
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
</script>

<div class="flex w-full flex-col gap-3">
	<div class="flex items-center gap-3">
		<span class="w-10 text-right font-mono text-xs text-muted tabular-nums">
			{formatDuration(Math.floor(player.currentTime))}
		</span>
		<RangeSlider
			label="Seek"
			value={player.currentTime}
			max={player.duration || 0}
			disabled={!player.track}
			oninput={(seconds) => player.seek(seconds)}
		/>
		<span class="w-10 font-mono text-xs text-muted tabular-nums">
			{formatDuration(Math.floor(player.duration))}
		</span>
	</div>

	{#if player.errorMessage}
		<p class="text-center text-xs text-red">{player.errorMessage}</p>
	{/if}

	<div class="flex items-center justify-center gap-3">
		<span
			title={shuffleLabel(player.queue)}
			class={[player.queue.shuffle && '[&>button]:text-action']}
		>
			<Button
				round
				variant="ghost"
				icon={shuffleIcon(player.queue)}
				onclick={() => player.cycleShuffle()}
			/>
		</span>
		<Button round variant="ghost" icon={SkipBackIcon} onclick={() => player.previous()} />
		{#if player.buffering && player.track && !player.errorMessage}
			<div class="flex size-14 items-center justify-center"><LoadingSpinner size={24} /></div>
		{:else}
			<PlayOverlayButton
				size="lg"
				playing={!player.paused}
				label={playButtonLabel()}
				onclick={() => player.toggle()}
			/>
		{/if}
		<Button round variant="ghost" icon={SkipForwardIcon} onclick={() => player.next()} />
		<span class={[player.queue.repeat !== 'off' && '[&>button]:text-action']}>
			<Button round variant="ghost" icon={RepeatButtonIcon} onclick={() => player.cycleRepeat()} />
		</span>
		<button
			type="button"
			aria-label={likeLabel()}
			aria-pressed={liked}
			onclick={toggleLikeOfCurrent}
			class={[
				'rounded-full p-2 hover:text-default',
				liked && 'text-action',
				!liked && 'text-muted'
			]}
		>
			<HeartIcon size={24} weight={likeWeight()} />
		</button>
	</div>
</div>
