<script lang="ts">
	import { Button, View } from '@neoworks/svelte-gpui/ui';
	import { podcast, podcastContinuation } from '#lib/api/catalog.ts';
	import { createItem, type Episode } from '#lib/api/music.ts';
	import EpisodeRow from '#lib/components/EpisodeRow.svelte';
	import FollowButton from '#lib/components/FollowButton.svelte';
	import Label from '#lib/components/Label.svelte';
	import LoadMoreSentinel from '#lib/components/LoadMoreSentinel.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import PageList from '#lib/components/PageList.svelte';
	import PageShell from '#lib/components/PageShell.svelte';
	import PageStatus from '#lib/components/PageStatus.svelte';
	import PlaylistActionRow from '#lib/components/PlaylistActionRow.svelte';
	import StickyPlayBar from '#lib/components/StickyPlayBar.svelte';
	import { itemHref } from '#lib/format.ts';
	import { ContinuationList } from '#lib/pages/continuationList.svelte.ts';
	import type { PageProps } from '#lib/pages/pageProps.ts';
	import { RemoteResource } from '#lib/pages/remoteResource.svelte.ts';
	import { routeParameter } from '#lib/pages/routeParameter.ts';
	import { playContext } from '#lib/player/contextPlayback.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { colors } from '#lib/theme.ts';

	type PodcastBlock =
		| { key: string; kind: 'header' }
		| { key: string; kind: 'actions' }
		| { key: string; kind: 'about' }
		| { key: string; kind: 'episodesTitle' }
		| { key: string; kind: 'episode'; index: number }
		| { key: string; kind: 'loadMore' };

	// The play button has scrolled out of view once the header and the action row are above the window.
	const stickyBarScrollOffset = 280;

	let { params }: PageProps = $props();

	const resource = new RemoteResource(podcast);
	const episodeList = new ContinuationList<Episode>(fetchMoreEpisodes);
	const podcastId = $derived(routeParameter(params.id));
	const podcastPage = $derived(resource.data);
	const context = $derived(`podcast:${podcastId}`);
	const episodeTracks = $derived(episodeList.items.map((episode) => episode.track));
	const blocks = $derived(buildBlocks());
	let scrollOffset = $state(0);

	$effect(() => {
		resource.load(podcastId);
		scrollOffset = 0;
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

	function buildBlocks(): PodcastBlock[] {
		const built: PodcastBlock[] = [
			{ key: 'header', kind: 'header' },
			{ key: 'actions', kind: 'actions' }
		];
		if (podcastPage?.description) {
			built.push({ key: 'about', kind: 'about' });
		}
		built.push({ key: 'episodesTitle', kind: 'episodesTitle' });
		episodeList.items.forEach((_, index) => {
			built.push({ key: `episode:${index}`, kind: 'episode', index });
		});
		if (episodeList.hasMore) {
			built.push({ key: `loadMore:${episodeList.items.length}`, kind: 'loadMore' });
		}
		return built;
	}

	function playEpisode(index: number): void {
		const track = episodeTracks[index];
		player.playOrToggle(track, episodeTracks, index, context, 'episode');
	}
</script>

<View style={{ flex: 1, minHeight: 0 }}>
	<PageStatus status={resource.status} errorMessage={resource.errorMessage} />

	{#if podcastPage}
		<PageShell seed={podcastPage.title}>
			<PageList
				items={blocks}
				key={(block) => block.key}
				paddingX={24}
				paddingTop={24}
				paddingBottom={24}
				estimatedRowHeight={128}
				onscroll={(position) => (scrollOffset = position.y)}
			>
				{#snippet row(block)}
					{#if block.kind === 'header'}
						<PageHeader
							label="Podcast"
							title={podcastPage.title}
							thumbnails={podcastPage.thumbnails}
						>
							{#snippet meta()}
								{#if podcastPage.author.id}
									<Button
										hover={{ underline: true }}
										onpress={() => goto(itemHref('artist', podcastPage.author.id))}
									>
										<Label size="small">{podcastPage.author.name}</Label>
									</Button>
								{:else}
									<Label size="small">{podcastPage.author.name}</Label>
								{/if}
							{/snippet}
						</PageHeader>
					{:else if block.kind === 'actions'}
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
					{:else if block.kind === 'about'}
						<View style={{ maxWidth: 768, gap: 8, paddingBottom: 8 }}>
							<Label size="heading" weight="bold">About</Label>
							<Label size="small" tone="muted" lines={4}>{podcastPage.description}</Label>
						</View>
					{:else if block.kind === 'episodesTitle'}
						<Label size="heading" weight="bold" style={{ paddingBottom: 8 }}>All episodes</Label>
					{:else if block.kind === 'episode'}
						<EpisodeRow
							episode={episodeList.items[block.index]}
							onplay={() => playEpisode(block.index)}
						/>
					{:else}
						<LoadMoreSentinel onvisible={() => episodeList.loadMore()} />
					{/if}
				{/snippet}
			</PageList>
			<StickyPlayBar
				visible={scrollOffset > stickyBarScrollOffset}
				{context}
				title={podcastPage.title}
				onplay={() => playContext(episodeTracks, context, 'episode')}
			/>
		</PageShell>
	{/if}
</View>
