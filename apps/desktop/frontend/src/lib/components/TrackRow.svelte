<script lang="ts">
	import { ListRow } from '@neoworks-dev/ui';
	import MusicNoteIcon from 'phosphor-svelte/lib/MusicNoteIcon';
	import SpeakerHighIcon from 'phosphor-svelte/lib/SpeakerHighIcon';
	import type { Track } from '#lib/api/music.ts';
	import { formatArtists, formatDuration, pickThumbnail } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';

	let { track }: { track: Track } = $props();

	const coverUrl = $derived(pickThumbnail(track.thumbnails, 96));
	const subtitle = $derived(buildSubtitle(track));
	const isPlaying = $derived(player.isCurrent(track));

	function buildSubtitle(track: Track): string {
		const artists = formatArtists(track.artists);
		if (!track.album) {
			return artists;
		}
		return `${artists} · ${track.album.name}`;
	}
</script>

<ListRow title={track.title} {subtitle} onclick={() => player.play(track)}>
	{#snippet leading()}
		{#if coverUrl}
			<img src={coverUrl} alt="" class="size-10 rounded-sm object-cover" loading="lazy" />
		{:else}
			<div class="flex size-10 items-center justify-center rounded-sm bg-hover text-faint">
				<MusicNoteIcon size={18} />
			</div>
		{/if}
	{/snippet}
	{#snippet trailing()}
		{#if isPlaying}
			<SpeakerHighIcon size={16} class="text-default" aria-label="Now playing" />
		{:else}
			<span class="font-mono text-xs text-dim tabular-nums">
				{formatDuration(track.durationSeconds)}
			</span>
		{/if}
	{/snippet}
</ListRow>
