<script lang="ts">
	import { page } from '$app/state';
	import { artist } from '#lib/api/catalog.ts';
	import { createItem } from '#lib/api/music.ts';
	import ArtistHero from '#lib/components/ArtistHero.svelte';
	import FollowButton from '#lib/components/FollowButton.svelte';
	import MediaCardRow from '#lib/components/MediaCardRow.svelte';
	import MoreMenuButton from '#lib/components/MoreMenuButton.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import TrackTable from '#lib/components/TrackTable.svelte';
	import { collectionMenuEntries } from '#lib/menu/collectionMenuEntries.ts';
	import { routeParameter } from '#lib/pages/routeParameter.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { playContext } from '#lib/player/contextPlayback.ts';

	const popularCollapsedCount = 5;
	const popularExpandedCount = 10;

	const resource = new RemoteResource(artist);
	const artistId = $derived(routeParameter(page.params.id));
	const artistPage = $derived(resource.data);
	const context = $derived(`artist:${artistId}`);
	let showMorePopular = $state(false);

	$effect(() => {
		resource.load(artistId);
		showMorePopular = false;
	});

	function popularCount(): number {
		if (showMorePopular) {
			return popularExpandedCount;
		}
		return popularCollapsedCount;
	}
</script>

<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

{#if artistPage}
	<ArtistHero name={artistPage.name} thumbnails={artistPage.thumbnails} />
	<div class="flex flex-col px-6 pb-6">
		<PlaylistActionRow
			{context}
			title={artistPage.name}
			onplay={() => playContext(artistPage.topSongs, context)}
		>
			<FollowButton
				item={createItem('artist', artistId, artistPage.name, 'Artist', artistPage.thumbnails)}
			/>
			<MoreMenuButton buildEntries={() => collectionMenuEntries(artistPage.topSongs)} />
		</PlaylistActionRow>

		{#if artistPage.topSongs.length > 0}
			<section class="flex flex-col gap-2 pb-8">
				<h2 class="text-2xl font-bold">Popular</h2>
				<TrackTable
					tracks={artistPage.topSongs}
					{context}
					showHeader={false}
					showAlbum={false}
					visibleCount={popularCount()}
				/>
				{#if artistPage.topSongs.length > popularCollapsedCount}
					<button
						type="button"
						onclick={() => (showMorePopular = !showMorePopular)}
						class="self-start px-4 text-sm font-semibold text-dim hover:text-default"
					>
						{#if showMorePopular}Show less{:else}See more{/if}
					</button>
				{/if}
			</section>
		{/if}

		<div class="-mx-3 flex flex-col gap-8">
			{#each artistPage.shelves as shelf, index (index)}
				<MediaCardRow title={shelf.title} items={shelf.items} />
			{/each}
		</div>
	</div>
{/if}
