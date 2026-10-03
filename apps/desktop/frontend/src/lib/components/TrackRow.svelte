<script lang="ts">
	import { ListRow } from '@neoworks-dev/ui';
	import MusicNoteIcon from 'phosphor-svelte/lib/MusicNoteIcon';
	import type { Track } from '#lib/api/music.ts';
	import { formatArtists, formatDuration, pickThumbnail } from '#lib/format.ts';

	let { track }: { track: Track } = $props();

	const coverUrl = $derived(pickThumbnail(track.thumbnails, 96));
	const subtitle = $derived(buildSubtitle(track));

	function buildSubtitle(track: Track): string {
		const artists = formatArtists(track.artists);
		if (!track.album) {
			return artists;
		}
		return `${artists} · ${track.album.name}`;
	}
</script>

<ListRow title={track.title} {subtitle}>
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
		<span class="font-mono text-xs text-dim tabular-nums"
			>{formatDuration(track.durationSeconds)}</span
		>
	{/snippet}
</ListRow>
