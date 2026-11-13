<script lang="ts">
	import { Button, View } from '@neoworks/svelte-gpui/ui';
	import { artist } from '#lib/api/catalog.ts';
	import { createItem, type Shelf } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import ArtistHero from '#lib/components/ArtistHero.svelte';
	import FollowButton from '#lib/components/FollowButton.svelte';
	import Label from '#lib/components/Label.svelte';
	import MediaCardRow from '#lib/components/MediaCardRow.svelte';
	import MoreMenuButton from '#lib/components/MoreMenuButton.svelte';
	import PageList from '#lib/components/PageList.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import StickyPlayBar from '#lib/components/StickyPlayBar.svelte';
	import TrackTable from '#lib/components/TrackTable.svelte';
	import { collectionMenuEntries } from '#lib/menu/collectionMenuEntries.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { routeParameter } from '#lib/pages/routeParameter.ts';
	import type { PageProps } from '#lib/pages/pageProps.ts';
	import { playContext } from '#lib/player/contextPlayback.ts';
	import { colors, gutter, hoverTransition, radii } from '#lib/theme.ts';

	type ArtistBlock =
		| { key: string; kind: 'hero' }
		| { key: string; kind: 'actions' }
		| { key: string; kind: 'popular' }
		| { key: string; kind: 'shelf'; shelf: Shelf };

	const popularCollapsedCount = 5;
	const popularExpandedCount = 10;
	// The play button has scrolled out of view once the hero and the action row are above the window.
	const stickyBarScrollOffset = 360;

	let { params }: PageProps = $props();

	const resource = new RemoteResource(artist);
	const artistId = $derived(routeParameter(params.id));
	const artistPage = $derived(resource.data);
	const context = $derived(`artist:${artistId}`);
	let showMorePopular = $state(false);
	let scrollOffset = $state(0);
	const blocks = $derived(buildBlocks());
	const popularCount = $derived(popularCountFor(showMorePopular));

	$effect(() => {
		resource.load(artistId);
		showMorePopular = false;
		scrollOffset = 0;
	});

	function popularCountFor(expanded: boolean): number {
		if (expanded) {
			return popularExpandedCount;
		}
		return popularCollapsedCount;
	}

	function buildBlocks(): ArtistBlock[] {
		if (!artistPage) {
			return [];
		}
		const built: ArtistBlock[] = [
			{ key: 'hero', kind: 'hero' },
			{ key: 'actions', kind: 'actions' }
		];
		if (artistPage.topSongs.length > 0) {
			built.push({ key: 'popular', kind: 'popular' });
		}
		artistPage.shelves.forEach((shelf, index) => {
			built.push({ key: `shelf:${index}`, kind: 'shelf', shelf });
		});
		return built;
	}
</script>

<View style={{ flex: 1, minHeight: 0, position: 'relative' }}>
	<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

	{#if artistPage}
		<PageList
			items={blocks}
			key={(block) => block.key}
			paddingX={0}
			paddingTop={0}
			paddingBottom={24}
			estimatedRowHeight={320}
			onscroll={(position) => (scrollOffset = position.y)}
		>
			{#snippet row(block)}
				{#if block.kind === 'hero'}
					<ArtistHero name={artistPage.name} thumbnails={artistPage.thumbnails} />
				{:else if block.kind === 'actions'}
					<View style={{ paddingX: gutter.wide }}>
						<PlaylistActionRow
							{context}
							title={artistPage.name}
							onplay={() => playContext(artistPage.topSongs, context)}
						>
							<FollowButton
								item={createItem(
									'artist',
									artistId,
									artistPage.name,
									'Artist',
									artistPage.thumbnails
								)}
							/>
							<MoreMenuButton buildEntries={() => collectionMenuEntries(artistPage.topSongs)} />
						</PlaylistActionRow>
					</View>
				{:else if block.kind === 'popular'}
					<View style={{ paddingX: gutter.wide, paddingBottom: 8, gap: 8 }}>
						<Label size="heading" weight="bold">Popular</Label>
						<TrackTable
							tracks={artistPage.topSongs}
							{context}
							showHeader={false}
							showAlbum={false}
							visibleCount={popularCount}
						/>
						{#if artistPage.topSongs.length > popularCollapsedCount}
							<Button
								style={{
									alignSelf: 'start',
									paddingX: 16,
									paddingY: 4,
									radius: radii.full,
									color: appearance.dim,
									transition: hoverTransition
								}}
								hover={{ color: colors.text }}
								onpress={() => (showMorePopular = !showMorePopular)}
							>
								<Label size="small" weight="semibold">
									{#if showMorePopular}Show less{:else}See more{/if}
								</Label>
							</Button>
						{/if}
					</View>
				{:else}
					<View style={{ paddingX: gutter.narrow, paddingTop: 32 }}>
						<MediaCardRow title={block.shelf.title} items={block.shelf.items} />
					</View>
				{/if}
			{/snippet}
		</PageList>
		<StickyPlayBar
			visible={scrollOffset > stickyBarScrollOffset}
			{context}
			title={artistPage.name}
			onplay={() => playContext(artistPage.topSongs, context)}
		/>
	{/if}
</View>
