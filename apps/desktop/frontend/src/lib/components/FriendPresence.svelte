<script lang="ts">
	import PauseIcon from 'phosphor-svelte/lib/PauseIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import type { FriendPresence } from '#lib/api/together.ts';
	import { formatDuration } from '#lib/format.ts';
	import { livePositionSeconds, listenTogether } from '#lib/together/listenTogether.svelte.ts';

	let { presence, userIdText }: { presence: FriendPresence | null; userIdText: string } = $props();

	const nowPlaying = $derived(presence?.nowPlaying);
	const artists = $derived(describeArtists());
	const position = $derived(describePosition());

	function describeArtists(): string {
		if (!nowPlaying) {
			return '';
		}
		return nowPlaying.track.artists.map((artist) => artist.name).join(', ');
	}

	function describePosition(): string {
		if (!nowPlaying) {
			return '';
		}
		const seconds = Math.floor(livePositionSeconds(nowPlaying, listenTogether.nowMs));
		const elapsed = formatDuration(seconds) || '0:00';
		const total = formatDuration(nowPlaying.track.durationSeconds);
		if (total === '') {
			return elapsed;
		}
		return `${elapsed} / ${total}`;
	}
</script>

{#if !presence}
	<p class="truncate font-mono text-xs text-dim">{userIdText}</p>
{:else if nowPlaying}
	<p class="flex min-w-0 items-center gap-1.5 text-xs text-muted">
		{#if nowPlaying.paused}
			<PauseIcon size={10} weight="fill" class="shrink-0" />
		{:else}
			<PlayIcon size={10} weight="fill" class="shrink-0 text-default" />
		{/if}
		<span class="truncate">
			<span class="text-default">{nowPlaying.track.title}</span>
			{#if artists !== ''}· {artists}{/if}
		</span>
		<span class="shrink-0 tabular-nums">{position}</span>
	</p>
{:else}
	<p class="text-xs text-muted">Online</p>
{/if}
