<script lang="ts">
	import { Button, LoadingSpinner } from '@neoworks-dev/ui';
	import MusicNoteIcon from 'phosphor-svelte/lib/MusicNoteIcon';
	import PauseIcon from 'phosphor-svelte/lib/PauseIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import { formatArtists, formatDuration, pickThumbnail } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';

	const coverUrl = $derived(currentCoverUrl());

	function currentCoverUrl() {
		if (!player.track) {
			return undefined;
		}
		return pickThumbnail(player.track.thumbnails, 120);
	}

	function handleSeek(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		player.seek(Number(input.value));
	}
</script>

<audio
	src={player.source}
	autoplay
	bind:paused={player.paused}
	bind:currentTime={player.currentTime}
	bind:duration={player.duration}
	onwaiting={() => (player.buffering = true)}
	onplaying={() => (player.buffering = false)}
	onerror={() => player.fail('This song could not be played.')}
></audio>

{#if player.track}
	<footer class="flex items-center gap-4 border-t border-line-faint bg-elevated px-6 py-3">
		<div class="flex w-72 min-w-0 items-center gap-3">
			{#if coverUrl}
				<img src={coverUrl} alt="" class="size-12 rounded-md object-cover" />
			{:else}
				<div class="flex size-12 items-center justify-center rounded-md bg-hover text-faint">
					<MusicNoteIcon size={20} />
				</div>
			{/if}
			<div class="min-w-0">
				<p class="truncate text-sm font-medium">{player.track.title}</p>
				<p class="truncate text-xs text-dim">{formatArtists(player.track.artists)}</p>
			</div>
		</div>

		<div class="flex flex-1 items-center gap-3">
			{#if player.buffering && !player.errorMessage}
				<div class="flex size-11 items-center justify-center"><LoadingSpinner size={20} /></div>
			{:else if player.paused}
				<Button round variant="primary" icon={PlayIcon} onclick={() => player.toggle()} />
			{:else}
				<Button round variant="primary" icon={PauseIcon} onclick={() => player.toggle()} />
			{/if}

			{#if player.errorMessage}
				<p class="text-sm text-red">{player.errorMessage}</p>
			{:else}
				<span class="w-12 text-right font-mono text-xs text-dim tabular-nums">
					{formatDuration(Math.floor(player.currentTime))}
				</span>
				<input
					type="range"
					aria-label="Seek"
					min="0"
					max={player.duration || 0}
					step="1"
					value={player.currentTime}
					oninput={handleSeek}
					class="flex-1 accent-(--primary)"
				/>
				<span class="w-12 font-mono text-xs text-dim tabular-nums">
					{formatDuration(Math.floor(player.duration))}
				</span>
			{/if}
		</div>
	</footer>
{/if}
