<script lang="ts">
	import { Button, View } from '@neoworks/svelte-gpui/ui';
	import { radio } from '#lib/api/catalog.ts';
	import { lyrics, type Lyrics } from '#lib/api/lyrics.ts';
	import type { Track } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import AddToPlaylistButton from '#lib/components/AddToPlaylistButton.svelte';
	import ArtistLinks from '#lib/components/ArtistLinks.svelte';
	import DownloadButton from '#lib/components/DownloadButton.svelte';
	import EmptyMessage from '#lib/components/EmptyMessage.svelte';
	import Label from '#lib/components/Label.svelte';
	import LikeButton from '#lib/components/LikeButton.svelte';
	import MoreMenuButton from '#lib/components/MoreMenuButton.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageList from '#lib/components/PageList.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import StickyPlayBar from '#lib/components/StickyPlayBar.svelte';
	import TrackListRow from '#lib/components/TrackListRow.svelte';
	import TrackTableHeader from '#lib/components/TrackTableHeader.svelte';
	import { trackRowHeight } from '#lib/components/trackColumns.ts';
	import { formatDuration, hasLinkedAlbums, itemHref, playCountOf } from '#lib/format.ts';
	import { lyricsSettings } from '#lib/lyrics/lyricsSettings.svelte.ts';
	import { buildTrackMenuEntries } from '#lib/menu/trackMenuEntries.ts';
	import type { PageProps } from '#lib/pages/pageProps.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { routeParameter } from '#lib/pages/routeParameter.ts';
	import { ensureTrackStates } from '#lib/pages/trackStates.ts';
	import { playContext } from '#lib/player/contextPlayback.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { colors } from '#lib/theme.ts';

	type SongBlock =
		| { key: string; kind: 'header' }
		| { key: string; kind: 'actions' }
		| { key: string; kind: 'lyrics' }
		| { key: string; kind: 'relatedTitle' }
		| { key: string; kind: 'tableHeader' }
		| { key: string; kind: 'related'; index: number };

	// Lines shown before "Show all lyrics".
	const lyricsPreviewLines = 8;
	// The play button has scrolled out of view once the header and the action row are above the window.
	const stickyBarScrollOffset = 280;

	let { params }: PageProps = $props();

	// The radio of a song starts with the song itself, then related songs.
	const resource = new RemoteResource(radio);
	const videoId = $derived(routeParameter(params.id));
	const radioTracks = $derived(loadedTracks(resource.data));
	const song = $derived(findSong(radioTracks));
	const related = $derived(radioTracks.filter((track) => track.videoId !== videoId));
	const context = $derived(`song:${videoId}`);
	const showAlbum = $derived(hasLinkedAlbums(related));
	const blocks = $derived(buildBlocks());

	let songLyrics = $state<Lyrics | null>(null);
	let lyricsExpanded = $state(false);
	let scrollOffset = $state(0);
	const shownLyricLines = $derived(visibleLyricLines());
	const lyricsToggleLabel = $derived(lyricsToggleLabelFor(lyricsExpanded));

	$effect(() => {
		resource.load(videoId);
		scrollOffset = 0;
	});

	$effect(() => {
		if (song) {
			loadLyrics(song, lyricsSettings.allowLRCLIB);
		}
	});

	$effect(() => {
		ensureTrackStates(related);
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

	function lyricsToggleLabelFor(expanded: boolean): string {
		if (expanded) {
			return 'Show less';
		}
		return 'Show all lyrics';
	}

	function buildBlocks(): SongBlock[] {
		const built: SongBlock[] = [
			{ key: 'header', kind: 'header' },
			{ key: 'actions', kind: 'actions' }
		];
		if (shownLyricLines.length > 0) {
			built.push({ key: 'lyrics', kind: 'lyrics' });
		}
		if (related.length > 0) {
			built.push({ key: 'relatedTitle', kind: 'relatedTitle' });
			built.push({ key: 'tableHeader', kind: 'tableHeader' });
		}
		related.forEach((_, index) => {
			built.push({ key: `related:${index}`, kind: 'related', index });
		});
		return built;
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

<View style={{ flex: 1, minHeight: 0 }}>
	<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

	{#if resource.status === 'done' && !song}
		<EmptyMessage>This song could not be found.</EmptyMessage>
	{/if}

	{#if song}
		<PageShell seed={song.title} cover={song.thumbnails}>
			<PageList
				items={blocks}
				key={(block) => block.key}
				paddingX={24}
				paddingTop={24}
				paddingBottom={24}
				estimatedRowHeight={trackRowHeight}
				onscroll={(position) => (scrollOffset = position.y)}
			>
				{#snippet row(block)}
					{#if block.kind === 'header'}
						<PageHeader label="Song" title={song.title} thumbnails={song.thumbnails}>
							{#snippet meta()}
								<ArtistLinks artists={song.artists} color={appearance.muted} />
								{#if song.album && song.album.id}
									<Label size="small">•</Label>
									<Button
										hover={{ color: colors.text, underline: true }}
										onpress={() => goto(itemHref('album', song.album!.id))}
									>
										<Label size="small">{song.album.name}</Label>
									</Button>
								{/if}
								{#if detailsOf(song) !== ''}
									<Label size="small">• {detailsOf(song)}</Label>
								{/if}
							{/snippet}
						</PageHeader>
					{:else if block.kind === 'actions'}
						<PlaylistActionRow {context} title={song.title} onplay={playSong}>
							<LikeButton track={song} size={28} />
							<AddToPlaylistButton track={song} />
							<DownloadButton track={song} size={24} revealOnHover={false} />
							<MoreMenuButton buildEntries={() => buildTrackMenuEntries(song)} />
						</PlaylistActionRow>
					{:else if block.kind === 'lyrics'}
						<View style={{ maxWidth: 672, gap: 12, paddingBottom: 32 }}>
							<Label size="heading" weight="bold">Lyrics</Label>
							<View style={{ gap: 4 }}>
								{#each shownLyricLines as line, index (index)}
									<Label size="large" weight="semibold" tone="muted">{line}</Label>
								{/each}
							</View>
							{#if hasMoreLyrics()}
								<Button
									style={{ alignSelf: 'start', color: appearance.dim }}
									hover={{ color: colors.text }}
									onpress={() => (lyricsExpanded = !lyricsExpanded)}
								>
									<Label size="small" weight="semibold">{lyricsToggleLabel}</Label>
								</Button>
							{/if}
							{#if songLyrics?.source}
								<Label size="tiny" tone="faint">{songLyrics.source}</Label>
							{/if}
						</View>
					{:else if block.kind === 'relatedTitle'}
						<Label size="heading" weight="bold" style={{ paddingBottom: 8 }}>Related songs</Label>
					{:else if block.kind === 'tableHeader'}
						<TrackTableHeader {showAlbum} />
					{:else}
						<TrackListRow
							tracks={related}
							index={block.index}
							context={`${context}:related`}
							{showAlbum}
						/>
					{/if}
				{/snippet}
			</PageList>
			<StickyPlayBar
				visible={scrollOffset > stickyBarScrollOffset}
				{context}
				title={song.title}
				onplay={playSong}
			/>
		</PageShell>
	{/if}
</View>
