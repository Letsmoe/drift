<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button, FloatingScrollbar } from '@neoworks-dev/ui';
	import BooksIcon from 'phosphor-svelte/lib/BooksIcon';
	import DownloadSimpleIcon from 'phosphor-svelte/lib/DownloadSimpleIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import ArrowCircleDownIcon from 'phosphor-svelte/lib/ArrowCircleDownIcon';
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import ImportStatusCard from '#lib/components/ImportStatusCard.svelte';
	import MusicImportDialog from '#lib/components/MusicImportDialog.svelte';
	import Chip from '#lib/components/Chip.svelte';
	import LibraryRow from '#lib/components/LibraryRow.svelte';
	import { listenToImport } from '#lib/api/musicImport.ts';
	import { musicImport } from '#lib/import/musicImport.svelte.ts';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { formatSongCount } from '#lib/format.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { importPrompt } from '#lib/library/importPrompt.svelte.ts';
	import { LibraryListing, libraryFilters } from '#lib/library/libraryListing.svelte.ts';

	const listing = new LibraryListing();
	const showPlaylistPrompt = $derived(
		library.playlists.length === 0 && (listing.filter === 'all' || listing.filter === 'playlist')
	);
	const showPodcastPrompt = $derived(
		!listing.hasKind('podcast') && (listing.filter === 'all' || listing.filter === 'podcast')
	);
	const showLikedSongs = $derived(listing.filter === 'all' || listing.filter === 'playlist');

	// Someone who has imported once knows where importing lives.
	$effect(() =>
		listenToImport({
			onProgress: () => {},
			onDone: () => importPrompt.hide(),
			onError: () => {}
		})
	);

	function downloadsSubtitle(): string {
		const active = downloads.summary.queued + downloads.summary.downloading;
		if (active > 0) {
			return `${active} downloading`;
		}
		return `${formatSongCount(downloads.summary.done)} offline`;
	}

	async function createPlaylist() {
		const created = await library.createNumberedPlaylist();
		await goto(`/library/playlist/${encodeURIComponent(created.id)}`);
	}

	function findPodcasts() {
		goto('/search?q=podcasts');
	}
</script>

<aside class="glass flex min-h-0 w-80 shrink-0 flex-col gap-2 rounded-2xl p-2">
	<header class="flex h-12 items-center gap-3 px-2 text-dim">
		<BooksIcon size={22} />
		<h2 class="flex-1 text-base font-semibold">Your Library</h2>
		<span title="Import your music" class="flex">
			<Button
				round
				size="sm"
				variant="ghost"
				icon={DownloadSimpleIcon}
				onclick={() => musicImport.openDialog()}
			/>
		</span>
		<span title="Create playlist" class="flex">
			<Button round size="sm" variant="ghost" icon={PlusIcon} onclick={createPlaylist} />
		</span>
	</header>

	<div class="flex flex-wrap gap-2 px-2 pb-1">
		{#each libraryFilters as { filter, label } (filter)}
			<Chip {label} selected={listing.filter === filter} onclick={() => listing.toggle(filter)} />
		{/each}
	</div>

	<FloatingScrollbar class="min-h-0 flex-1">
		<div class="flex flex-col">
			{#if showLikedSongs}
				<LibraryRow
					href="/collection/tracks"
					title="Liked Songs"
					subtitle="Playlist"
					context="liked"
				>
					{#snippet cover()}
						<div
							class="flex size-12 shrink-0 items-center justify-center rounded-md bg-hover text-default"
						>
							<HeartIcon size={20} weight="fill" />
						</div>
					{/snippet}
				</LibraryRow>
			{/if}
			{#if showLikedSongs}
				<LibraryRow
					href="/downloads"
					title="Downloads"
					subtitle={downloadsSubtitle()}
					context="downloads"
				>
					{#snippet cover()}
						<div
							class="flex size-12 shrink-0 items-center justify-center rounded-md bg-hover text-default"
						>
							<ArrowCircleDownIcon size={22} weight="fill" />
						</div>
					{/snippet}
				</LibraryRow>
			{/if}
			{#each listing.entries as entry (entry.key)}
				<LibraryRow
					href={entry.href}
					title={entry.title}
					subtitle={entry.subtitle}
					context={entry.context}
					thumbnails={entry.thumbnails}
					round={entry.round}
				/>
			{/each}
		</div>

		<div class="flex flex-col gap-3 p-2">
			{#if showPlaylistPrompt}
				<div class="flex flex-col gap-3 rounded-lg bg-raised p-4">
					<div>
						<p class="text-base font-semibold">Create your first playlist</p>
						<p class="text-sm text-muted">It's easy, we'll help you.</p>
					</div>
					<div class="self-start">
						<Button variant="primary" size="sm" onclick={createPlaylist}>Create playlist</Button>
					</div>
				</div>
			{/if}
			{#if showPodcastPrompt}
				<div class="flex flex-col gap-3 rounded-lg bg-raised p-4">
					<div>
						<p class="text-base font-semibold">Let's find some podcasts to follow</p>
						<p class="text-sm text-muted">We'll keep you updated on new episodes.</p>
					</div>
					<div class="self-start">
						<Button variant="primary" size="sm" onclick={findPodcasts}>Browse podcasts</Button>
					</div>
				</div>
			{/if}
			{#if !importPrompt.hidden}
				<div class="relative flex flex-col gap-3 rounded-lg bg-raised p-4">
					<button
						type="button"
						aria-label="Dismiss"
						title="Dismiss"
						onclick={() => importPrompt.hide()}
						class="absolute top-2 right-2 rounded-full p-1.5 text-dim hover:bg-hover hover:text-default"
					>
						<XIcon size={14} />
					</button>
					<div class="pr-6">
						<p class="text-base font-semibold">Import your music</p>
						<p class="text-sm text-muted">
							From YouTube Music or Spotify: liked songs, playlists, albums and more.
						</p>
					</div>
					<div class="self-start">
						<Button variant="primary" size="sm" onclick={() => musicImport.openDialog()}>
							Import your music
						</Button>
					</div>
				</div>
			{/if}
		</div>
	</FloatingScrollbar>

	{#if musicImport.hasBackgroundStatus && !musicImport.dialogOpen}
		<ImportStatusCard flow={musicImport} />
	{/if}
</aside>

{#if musicImport.dialogOpen}
	<MusicImportDialog />
{/if}
