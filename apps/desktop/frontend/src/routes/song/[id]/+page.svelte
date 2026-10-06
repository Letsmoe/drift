<script lang="ts">
	import { page } from '$app/state';
	import { radio } from '#lib/api/catalog.ts';
	import { lyrics, type Lyrics } from '#lib/api/lyrics.ts';
	import type { Track } from '#lib/api/music.ts';
	import AddToPlaylistButton from '#lib/components/AddToPlaylistButton.svelte';
	import ArtistLinks from '#lib/components/ArtistLinks.svelte';
	import DownloadButton from '#lib/components/DownloadButton.svelte';
	import LikeButton from '#lib/components/LikeButton.svelte';
	import MoreMenuButton from '#lib/components/MoreMenuButton.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import TrackTable from '#lib/components/TrackTable.svelte';
	import { formatDuration, itemHref, playCountOf } from '#lib/format.ts';
	import { lyricsSettings } from '#lib/lyrics/lyricsSettings.svelte.ts';
	import { buildTrackMenuEntries } from '#lib/menu/trackMenuEntries.ts';
	import { routeParameter } from '#lib/pages/routeParameter.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { playContext } from '#lib/player/contextPlayback.ts';

	// Lines shown before "Show all lyrics".
	const lyricsPreviewLines = 8;

	// The radio of a song starts with the song itself, then related songs.
	const resource = new RemoteResource(radio);
	const videoId = $derived(routeParameter(page.params.id));
	const radioTracks = $derived(loadedTracks(resource.data));
	const song = $derived(findSong(radioTracks));
	const related = $derived(radioTracks.filter((track) => track.videoId !== videoId));
	const context = $derived(`song:${videoId}`);

	let songLyrics = $state<Lyrics | null>(null);
	let lyricsExpanded = $state(false);
	const shownLyricLines = $derived(visibleLyricLines());

	$effect(() => {
		resource.load(videoId);
	});

	$effect(() => {
		if (song) {
			loadLyrics(song, lyricsSettings.allowLRCLIB);
		}
	});

	function loadedTracks(data: Track[] | undefined): Track[] {
		if (!data) {
			return [];
		}
		return data;
	}

	function findSong(tracks: Track[]): Track | undefined {
		return tracks.find((track) => track.videoId === videoId);
	}

	async function loadLyrics(forSong: Track, allowLRCLIB: boolean): Promise<void> {
		songLyrics = null;
		lyricsExpanded = false;
		try {
			const found = await lyrics(forSong, allowLRCLIB);
			if (forSong.videoId === videoId) {
				songLyrics = found;
			}
		} catch {
			// Lyrics are optional; the page works without them.
		}
	}

	function visibleLyricLines(): string[] {
		if (!songLyrics) {
			return [];
		}
		const lines = songLyrics.lines.map((line) => line.text);
		if (lyricsExpanded) {
			return lines;
		}
		return lines.slice(0, lyricsPreviewLines);
	}

	function hasMoreLyrics(): boolean {
		return songLyrics !== null && songLyrics.lines.length > lyricsPreviewLines;
	}

	// Playing the page plays the song, then its radio.
	function playSong(): void {
		if (!song) {
			return;
		}
		playContext([song, ...related], context);
	}

	function detailsOf(shown: Track): string {
		return [formatDuration(shown.durationSeconds), playCountOf(shown)]
			.filter((part) => part !== '')
			.join(' • ');
	}
</script>

<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

{#if resource.status === 'done' && !song}
	<p class="py-24 text-center text-sm text-dim">This song could not be found.</p>
{/if}

{#if song}
	<PageShell seed={song.title}>
		<PageHeader label="Song" title={song.title} thumbnails={song.thumbnails}>
			{#snippet meta()}
				<ArtistLinks artists={song.artists} />
				{#if song.album && song.album.id}
					<span>•</span>
					<a href={itemHref('album', song.album.id)} class="hover:text-default hover:underline">
						{song.album.name}
					</a>
				{/if}
				{#if detailsOf(song) !== ''}
					<span>• {detailsOf(song)}</span>
				{/if}
			{/snippet}
		</PageHeader>
		<PlaylistActionRow {context} title={song.title} onplay={playSong}>
			<LikeButton track={song} size={28} />
			<AddToPlaylistButton track={song} />
			<span class="flex size-8 items-center justify-center text-muted">
				<DownloadButton track={song} size={24} revealOnHover={false} />
			</span>
			<MoreMenuButton buildEntries={() => buildTrackMenuEntries(song)} />
		</PlaylistActionRow>

		{#if shownLyricLines.length > 0}
			<section class="flex max-w-2xl flex-col gap-3 pb-8">
				<h2 class="text-2xl font-bold">Lyrics</h2>
				<div class="flex flex-col gap-1 text-lg font-semibold text-muted">
					{#each shownLyricLines as line, index (index)}
						<p>{line}</p>
					{/each}
				</div>
				{#if hasMoreLyrics()}
					<button
						type="button"
						onclick={() => (lyricsExpanded = !lyricsExpanded)}
						class="self-start text-sm font-semibold text-dim hover:text-default"
					>
						{#if lyricsExpanded}Show less{:else}Show all lyrics{/if}
					</button>
				{/if}
				{#if songLyrics?.source}
					<p class="text-xs text-faint">{songLyrics.source}</p>
				{/if}
			</section>
		{/if}

		{#if related.length > 0}
			<h2 class="pb-2 text-2xl font-bold">Related songs</h2>
			<TrackTable tracks={related} context="{context}:related" />
		{/if}
	</PageShell>
{/if}
