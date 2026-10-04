<script lang="ts">
	import { Button } from '@neoworks-dev/ui';
	import ArrowCircleDownIcon from 'phosphor-svelte/lib/ArrowCircleDownIcon';
	import PauseIcon from 'phosphor-svelte/lib/PauseIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import TrashIcon from 'phosphor-svelte/lib/TrashIcon';
	import type { DownloadCollectionInfo } from '#lib/api/downloads.ts';
	import type { Track } from '#lib/api/music.ts';
	import ConfirmDialog from '#lib/components/ConfirmDialog.svelte';
	import DownloadQueueRow from '#lib/components/DownloadQueueRow.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import TrackTable from '#lib/components/TrackTable.svelte';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { formatBytes } from '#lib/downloads/downloadProgress.ts';
	import { formatSongCount } from '#lib/format.ts';
	import { playContext } from '#lib/player/contextPlayback.ts';

	const context = 'downloads';
	const tracks = $derived<Track[]>(downloads.downloadedItems.map((item) => item.track));
	const paused = $derived(downloads.summary.paused);
	let confirmingDeleteAll = $state(false);

	$effect(() => downloads.watchLists());

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

<PageShell seed="Downloads">
	<PageHeader label="Library" title="Downloads">
		{#snippet cover()}
			<div
				class="flex size-full items-center justify-center bg-linear-to-br from-green to-blue text-white"
			>
				<ArrowCircleDownIcon size={96} weight="fill" />
			</div>
		{/snippet}
		{#snippet meta()}
			<span>{formatSongCount(downloads.usage.count)} • {formatBytes(downloads.usage.bytes)}</span>
		{/snippet}
	</PageHeader>

	<PlaylistActionRow {context} title="Downloads" onplay={() => playContext(tracks, context)}>
		{#if paused}
			<Button size="sm" icon={PlayIcon} onclick={() => downloads.resume()}>Resume downloads</Button>
		{:else}
			<Button size="sm" icon={PauseIcon} onclick={() => downloads.pause()}>Pause downloads</Button>
		{/if}
		<Button
			size="sm"
			variant="ghost"
			icon={TrashIcon}
			disabled={downloads.usage.count === 0 && downloads.queueItems.length === 0}
			onclick={() => (confirmingDeleteAll = true)}
		>
			Delete all downloads
		</Button>
	</PlaylistActionRow>

	{#if downloads.summary.blocked}
		<p class="pb-4 text-sm text-dim">Downloads are on hold until the connection allows them.</p>
	{/if}
	{#if downloads.errorMessage}
		<p class="pb-4 text-sm text-red">{downloads.errorMessage}</p>
	{/if}

	{#if downloads.queueItems.length > 0}
		<section class="pb-8">
			<h2 class="px-4 pb-2 text-lg font-semibold">Downloading</h2>
			{#each downloads.queueItems as item (item.track.videoId)}
				<DownloadQueueRow {item} />
			{/each}
		</section>
	{/if}

	{#if downloads.collections.length > 0}
		<section class="pb-8">
			<h2 class="px-4 pb-2 text-lg font-semibold">Collections</h2>
			{#each downloads.collections as collection (`${collection.kind}:${collection.id}`)}
				<a
					href={collectionHref(collection)}
					class="duration-fast flex flex-col rounded-md px-4 py-2 transition-colors hover:bg-hover"
				>
					<span class="truncate text-base">{collection.title}</span>
					<span class="truncate text-sm text-dim">{collectionSubtitle(collection)}</span>
				</a>
			{/each}
		</section>
	{/if}

	<section>
		<h2 class="px-4 pb-2 text-lg font-semibold">Downloaded songs</h2>
		{#if tracks.length === 0}
			<p class="py-12 text-center text-sm text-dim">
				Songs you download play without a connection and show up here.
			</p>
		{:else}
			<TrackTable {tracks} {context} />
		{/if}
	</section>
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
