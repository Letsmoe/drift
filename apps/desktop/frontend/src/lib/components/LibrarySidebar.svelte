<script lang="ts">
	import { Icon, Scroll, View } from '@neoworks/svelte-gpui/ui';
	import { Button } from '@neoworks/svelte-gpui-ui';
	import { listenToImport } from '#lib/api/musicImport.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Chip from '#lib/components/Chip.svelte';
	import IconAction from '#lib/components/IconAction.svelte';
	import Label from '#lib/components/Label.svelte';
	import LibraryRow from '#lib/components/LibraryRow.svelte';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { formatSongCount } from '#lib/format.ts';
	import { musicImport } from '#lib/import/musicImport.svelte.ts';
	import {
		arrowCircleDownFill,
		books,
		downloadSimple,
		heartFill,
		plus,
		x
	} from '#lib/layout/chromeIcons.ts';
	import { importPrompt } from '#lib/library/importPrompt.svelte.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { LibraryListing, libraryFilters } from '#lib/library/libraryListing.svelte.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { colors, glassBackground, radii } from '#lib/theme.ts';

	const sidebarWidth = 320;
	const tileSize = 48;

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

	async function createPlaylist(): Promise<void> {
		const created = await library.createNumberedPlaylist();
		goto(`/library/playlist/${encodeURIComponent(created.id)}`);
	}

	function findPodcasts(): void {
		goto('/search?q=podcasts');
	}
</script>

{#snippet tile(icon: string, iconSize: number)}
	<View
		style={{
			size: tileSize,
			shrink: 0,
			radius: radii.medium,
			background: colors.hover,
			align: 'center',
			justify: 'center',
			color: colors.text
		}}
	>
		<Icon svg={icon} size={iconSize} />
	</View>
{/snippet}

{#snippet promptCard(title: string, text: string, action: string, onaction: () => void)}
	<View style={{ gap: 12, padding: 16, radius: radii.large, background: colors.raised }}>
		<View>
			<Label size="medium" weight="semibold" tone="default">{title}</Label>
			<Label size="small" tone="muted">{text}</Label>
		</View>
		<View style={{ alignSelf: 'start' }}>
			<Button variant="primary" size="sm" onpress={onaction}>{action}</Button>
		</View>
	</View>
{/snippet}

<View
	style={{
		width: sidebarWidth,
		shrink: 0,
		minHeight: 0,
		gap: 8,
		padding: 8,
		radius: radii.extraLarge,
		background: glassBackground
	}}
>
	<View style={{ direction: 'row', align: 'center', gap: 12, height: 48, paddingX: 8 }}>
		<Icon svg={books} size={22} color={appearance.dim} />
		<View style={{ flex: 1 }}>
			<Label size="medium" weight="semibold" tone="dim">Your Library</Label>
		</View>
		<IconAction
			icon={downloadSimple}
			iconSize={18}
			boxSize={32}
			color={appearance.dim}
			hoverBackground={colors.hover}
			onpress={() => musicImport.openDialog()}
		/>
		<IconAction
			icon={plus}
			iconSize={18}
			boxSize={32}
			color={appearance.dim}
			hoverBackground={colors.hover}
			onpress={createPlaylist}
		/>
	</View>

	<View style={{ direction: 'row', wrap: true, gap: 8, paddingX: 8, paddingBottom: 4 }}>
		{#each libraryFilters as { filter, label } (filter)}
			<Chip {label} selected={listing.filter === filter} onclick={() => listing.toggle(filter)} />
		{/each}
	</View>

	<Scroll style={{ flex: 1, minHeight: 0 }}>
		{#if showLikedSongs}
			<LibraryRow href="/collection/tracks" title="Liked Songs" subtitle="Playlist" context="liked">
				{#snippet cover()}
					{@render tile(heartFill, 20)}
				{/snippet}
			</LibraryRow>
			<LibraryRow
				href="/downloads"
				title="Downloads"
				subtitle={downloadsSubtitle()}
				context="downloads"
			>
				{#snippet cover()}
					{@render tile(arrowCircleDownFill, 22)}
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

		<View style={{ gap: 12, padding: 8 }}>
			{#if showPlaylistPrompt}
				{@render promptCard(
					'Create your first playlist',
					"It's easy, we'll help you.",
					'Create playlist',
					createPlaylist
				)}
			{/if}
			{#if showPodcastPrompt}
				{@render promptCard(
					"Let's find some podcasts to follow",
					"We'll keep you updated on new episodes.",
					'Browse podcasts',
					findPodcasts
				)}
			{/if}
			{#if !importPrompt.hidden}
				<View style={{ gap: 12, padding: 16, radius: radii.large, background: colors.raised }}>
					<View style={{ position: 'absolute', top: 8, right: 8 }}>
						<IconAction
							icon={x}
							iconSize={14}
							boxSize={28}
							color={appearance.dim}
							hoverBackground={colors.hover}
							onpress={() => importPrompt.hide()}
						/>
					</View>
					<View style={{ paddingRight: 24 }}>
						<Label size="medium" weight="semibold" tone="default">Import your music</Label>
						<Label size="small" tone="muted">
							From YouTube Music or Spotify: liked songs, playlists, albums and more.
						</Label>
					</View>
					<View style={{ alignSelf: 'start' }}>
						<Button variant="primary" size="sm" onpress={() => musicImport.openDialog()}>
							Import your music
						</Button>
					</View>
				</View>
			{/if}
		</View>
	</Scroll>
</View>
