<script lang="ts">
	import { Icon, View } from '@neoworks/svelte-gpui/ui';
	import { Button } from '@neoworks/svelte-gpui-ui';
	import type { DownloadCollectionInfo } from '#lib/api/downloads.ts';
	import type { DownloadItem } from '#lib/api/downloads.ts';
	import type { Track } from '#lib/api/music.ts';
	import ConfirmDialog from '#lib/components/ConfirmDialog.svelte';
	import DownloadQueueRow from '#lib/components/DownloadQueueRow.svelte';
	import EmptyMessage from '#lib/components/EmptyMessage.svelte';
	import Label from '#lib/components/Label.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageList from '#lib/components/PageList.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import StickyPlayBar from '#lib/components/StickyPlayBar.svelte';
	import TrackListRow from '#lib/components/TrackListRow.svelte';
	import TrackTableHeader from '#lib/components/TrackTableHeader.svelte';
	import { trackRowHeight } from '#lib/components/trackColumns.ts';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { formatBytes } from '#lib/downloads/downloadProgress.ts';
	import { formatSongCount, hasLinkedAlbums } from '#lib/format.ts';
	import { arrowCircleDownFill, pause, play, trash } from '#lib/icons.ts';
	import { ensureTrackStates } from '#lib/pages/trackStates.ts';
	import { playContext, shufflePlayContext } from '#lib/player/contextPlayback.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	type DownloadsBlock =
		| { key: string; kind: 'header' }
		| { key: string; kind: 'actions' }
		| { key: string; kind: 'blocked' }
		| { key: string; kind: 'error' }
		| { key: string; kind: 'sectionTitle'; title: string }
		| { key: string; kind: 'queue'; item: DownloadItem }
		| { key: string; kind: 'collection'; collection: DownloadCollectionInfo }
		| { key: string; kind: 'empty' }
		| { key: string; kind: 'tableHeader' }
		| { key: string; kind: 'track'; index: number };

	const context = 'downloads';
	// The play button has scrolled out of view once the header and the action row are above the window.
	const stickyBarScrollOffset = 280;

	const tracks = $derived<Track[]>(downloads.downloadedItems.map((item) => item.track));
	const paused = $derived(downloads.summary.paused);
	const showAlbum = $derived(hasLinkedAlbums(tracks));
	const blocks = $derived(buildBlocks());
	let confirmingDeleteAll = $state(false);
	let scrollOffset = $state(0);

	$effect(() => downloads.watchLists());

	$effect(() => {
		ensureTrackStates(tracks);
	});

	function buildBlocks(): DownloadsBlock[] {
		return [
			{ key: 'header', kind: 'header' },
			{ key: 'actions', kind: 'actions' },
			...noticeBlocks(),
			...queueBlocks(),
			...collectionBlocks(),
			...songBlocks()
		];
	}

	function noticeBlocks(): DownloadsBlock[] {
		const built: DownloadsBlock[] = [];
		if (downloads.summary.blocked) {
			built.push({ key: 'blocked', kind: 'blocked' });
		}
		if (downloads.errorMessage) {
			built.push({ key: 'error', kind: 'error' });
		}
		return built;
	}

	function queueBlocks(): DownloadsBlock[] {
		if (downloads.queueItems.length === 0) {
			return [];
		}
		const built: DownloadsBlock[] = [
			{ key: 'queueTitle', kind: 'sectionTitle', title: 'Downloading' }
		];
		for (const item of downloads.queueItems) {
			built.push({ key: `queue:${item.track.videoId}`, kind: 'queue', item });
		}
		return built;
	}

	function collectionBlocks(): DownloadsBlock[] {
		if (downloads.collections.length === 0) {
			return [];
		}
		const built: DownloadsBlock[] = [
			{ key: 'collectionsTitle', kind: 'sectionTitle', title: 'Collections' }
		];
		for (const collection of downloads.collections) {
			built.push({
				key: `collection:${collection.kind}:${collection.id}`,
				kind: 'collection',
				collection
			});
		}
		return built;
	}

	function songBlocks(): DownloadsBlock[] {
		const built: DownloadsBlock[] = [
			{ key: 'songsTitle', kind: 'sectionTitle', title: 'Downloaded songs' }
		];
		if (tracks.length === 0) {
			built.push({ key: 'empty', kind: 'empty' });
			return built;
		}
		built.push({ key: 'tableHeader', kind: 'tableHeader' });
		tracks.forEach((track, index) => {
			built.push({ key: `track:${track.videoId}`, kind: 'track', index });
		});
		return built;
	}

	function collectionSubtitle(collection: DownloadCollectionInfo): string {
		const progress = `${collection.doneCount} of ${formatSongCount(collection.trackCount)}`;
		if (collection.keep) {
			return `${progress} • kept up to date`;
		}
		return progress;
	}

	function collectionHref(collection: DownloadCollectionInfo): string {
		switch (collection.kind) {
			case 'album':
				return `/album/${encodeURIComponent(collection.id)}`;
			case 'playlist':
				return `/playlist/${encodeURIComponent(collection.id)}`;
			case 'local':
				return `/library/playlist/${encodeURIComponent(collection.id)}`;
			default:
				return '/collection/tracks';
		}
	}
</script>

<View style={{ flex: 1, minHeight: 0 }}>
	<PageShell seed="Downloads">
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
					<PageHeader label="Library" title="Downloads">
						{#snippet cover()}
							<View
								style={{
									size: 'full',
									align: 'center',
									justify: 'center',
									background: colors.hover,
									color: colors.text
								}}
							>
								<Icon svg={arrowCircleDownFill} size={96} />
							</View>
						{/snippet}
						{#snippet meta()}
							<Label size="small">
								{formatSongCount(downloads.usage.count)} • {formatBytes(downloads.usage.bytes)}
							</Label>
						{/snippet}
					</PageHeader>
				{:else if block.kind === 'actions'}
					<PlaylistActionRow
						{context}
						title="Downloads"
						onplay={() => playContext(tracks, context)}
						onshuffle={() => shufflePlayContext(tracks, context)}
					>
						{#if paused}
							<Button size="sm" icon={play} onpress={() => downloads.resume()}>
								Resume downloads
							</Button>
						{:else}
							<Button size="sm" icon={pause} onpress={() => downloads.pause()}>
								Pause downloads
							</Button>
						{/if}
						<Button
							size="sm"
							variant="ghost"
							icon={trash}
							disabled={downloads.usage.count === 0 && downloads.queueItems.length === 0}
							onpress={() => (confirmingDeleteAll = true)}
						>
							Delete all downloads
						</Button>
					</PlaylistActionRow>
				{:else if block.kind === 'blocked'}
					<Label size="small" tone="dim" style={{ paddingBottom: 16 }}>
						Downloads are on hold until the connection allows them.
					</Label>
				{:else if block.kind === 'error'}
					<Label size="small" tone="danger" style={{ paddingBottom: 16 }}>
						{downloads.errorMessage}
					</Label>
				{:else if block.kind === 'sectionTitle'}
					<Label size="large" weight="semibold" style={{ paddingX: 16, paddingY: 8 }}>
						{block.title}
					</Label>
				{:else if block.kind === 'queue'}
					<DownloadQueueRow item={block.item} />
				{:else if block.kind === 'collection'}
					<View
						style={{
							paddingX: 16,
							paddingY: 8,
							radius: radii.medium,
							cursor: 'pointer',
							transition: hoverTransition
						}}
						hover={{ background: colors.hover }}
						onpress={() => goto(collectionHref(block.collection))}
					>
						<Label size="body" truncate>{block.collection.title}</Label>
						<Label size="small" tone="dim" truncate>
							{collectionSubtitle(block.collection)}
						</Label>
					</View>
				{:else if block.kind === 'empty'}
					<EmptyMessage>Songs you download play without a connection and show up here.</EmptyMessage
					>
				{:else if block.kind === 'tableHeader'}
					<TrackTableHeader {showAlbum} />
				{:else}
					<TrackListRow {tracks} index={block.index} {context} {showAlbum} />
				{/if}
			{/snippet}
		</PageList>
		<StickyPlayBar
			visible={scrollOffset > stickyBarScrollOffset}
			{context}
			title="Downloads"
			onplay={() => playContext(tracks, context)}
		/>
	</PageShell>

	{#if confirmingDeleteAll}
		<ConfirmDialog
			title="Delete all downloads"
			message="Delete every downloaded song from this device? Your library and playlists are not affected."
			confirmLabel="Delete all"
			onconfirm={() => downloads.deleteAll()}
			onclose={() => (confirmingDeleteAll = false)}
		/>
	{/if}
</View>
