<script lang="ts">
	import { page } from '$app/state';
	import { podcast, podcastContinuation } from '#lib/api/catalog.ts';
	import { createItem, type Episode } from '#lib/api/music.ts';
	import EpisodeRow from '#lib/components/EpisodeRow.svelte';
	import FollowButton from '#lib/components/FollowButton.svelte';
	import LoadMoreSentinel from '#lib/components/LoadMoreSentinel.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import { itemHref } from '#lib/format.ts';
	import { ContinuationList } from '#lib/pages/continuationList.svelte.ts';
	import { routeParameter } from '#lib/pages/routeParameter.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { playContext } from '#lib/player/contextPlayback.ts';
	import { player } from '#lib/player/player.svelte.ts';

	const resource = new RemoteResource(podcast);
	const episodeList = new ContinuationList<Episode>(fetchMoreEpisodes);
	const podcastId = $derived(routeParameter(page.params.id));
	const podcastPage = $derived(resource.data);
	const context = $derived(`podcast:${podcastId}`);
	const episodeTracks = $derived(episodeList.items.map((episode) => episode.track));

	$effect(() => {
		resource.load(podcastId);
	});

	$effect(() => {
		if (podcastPage) {
			episodeList.reset(podcastPage.episodes, podcastPage.continuation);
		}
	});

	async function fetchMoreEpisodes(token: string) {
		const next = await podcastContinuation(token);
		return { items: next.episodes, continuation: next.continuation };
	}

	function playEpisode(index: number) {
		const track = episodeTracks[index];
		player.playOrToggle(track, episodeTracks, index, context, 'episode');
	}
</script>

<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

{#if podcastPage}
	<PageShell seed={podcastPage.title}>
		<PageHeader label="Podcast" title={podcastPage.title} thumbnails={podcastPage.thumbnails}>
			{#snippet meta()}
				{#if podcastPage.author.id}
					<a href={itemHref('artist', podcastPage.author.id)} class="hover:underline">
						{podcastPage.author.name}
					</a>
				{:else}
					<span>{podcastPage.author.name}</span>
				{/if}
			{/snippet}
		</PageHeader>
		<PlaylistActionRow
			{context}
			title={podcastPage.title}
			onplay={() => playContext(episodeTracks, context, 'episode')}
		>
			<FollowButton
				item={createItem(
					'podcast',
					podcastId,
					podcastPage.title,
					podcastPage.author.name,
					podcastPage.thumbnails
				)}
			/>
		</PlaylistActionRow>
		{#if podcastPage.description}
			<section class="max-w-3xl pb-6">
				<h2 class="pb-2 text-2xl font-bold">About</h2>
				<p class="line-clamp-4 text-sm text-muted">{podcastPage.description}</p>
			</section>
		{/if}
		<h2 class="pb-2 text-2xl font-bold">All episodes</h2>
		<div class="flex flex-col">
			{#each episodeList.items as episode, index (index)}
				<EpisodeRow {episode} onplay={() => playEpisode(index)} />
			{/each}
		</div>
		{#if episodeList.hasMore}
			<LoadMoreSentinel
				itemCount={episodeList.items.length}
				onvisible={() => episodeList.loadMore()}
			/>
		{/if}
	</PageShell>
{/if}
