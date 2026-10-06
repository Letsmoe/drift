<script lang="ts">
	import { podcast } from '#lib/api/catalog.ts';
	import type { Item, PodcastPage } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { formatDuration, itemHref } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';

	let { item }: { item: Item } = $props();

	const latestEpisodeCount = 3;

	let page = $state<PodcastPage | null>(null);
	const episodes = $derived(latestEpisodes(page));
	const episodeTracks = $derived(episodes.map((episode) => episode.track));
	const context = $derived(`podcast:${item.id}`);

	$effect(() => {
		loadPage(item.id);
	});

	async function loadPage(showId: string): Promise<void> {
		page = null;
		try {
			const loaded = await podcast(showId);
			if (showId === item.id) {
				page = loaded;
			}
		} catch {
			// The card still shows the show without its episodes.
		}
	}

	function latestEpisodes(loaded: PodcastPage | null) {
		if (!loaded) {
			return [];
		}
		return loaded.episodes.slice(0, latestEpisodeCount);
	}

	function playEpisode(index: number): void {
		player.playOrToggle(episodeTracks[index], episodeTracks, index, context, 'episode');
	}
</script>

<section class="flex gap-6 rounded-2xl bg-raised p-5">
	<a href={itemHref('podcast', item.id)} class="shrink-0">
		<Cover thumbnails={item.thumbnails} size={208} class="size-52 rounded-xl shadow-lg" />
	</a>
	<div class="flex min-w-0 flex-1 flex-col gap-2">
		<p class="text-xs font-semibold tracking-wide text-dim uppercase">Featured podcast</p>
		<a href={itemHref('podcast', item.id)} class="text-3xl leading-tight font-bold hover:underline">
			{item.title}
		</a>
		<p class="text-sm text-muted">{item.subtitle}</p>
		{#if item.description}
			<p class="line-clamp-3 max-w-3xl text-sm text-dim">{item.description}</p>
		{/if}
		{#if episodes.length > 0}
			<div class="mt-2 flex flex-col">
				<p class="pb-1 text-xs font-semibold text-muted">Latest episodes</p>
				{#each episodes as episode, index (episode.track.videoId)}
					<div class="flex items-center gap-3 rounded-md px-2 py-1.5 hover:bg-hover">
						<PlayOverlayButton
							size="sm"
							playing={player.isPlaying(episode.track)}
							label="Play {episode.track.title}"
							onclick={() => playEpisode(index)}
						/>
						<p
							class={[
								'min-w-0 flex-1 truncate text-sm',
								player.isCurrent(episode.track) && 'text-action'
							]}
						>
							{episode.track.title}
						</p>
						<span class="shrink-0 text-xs text-dim">
							{[episode.published, formatDuration(episode.track.durationSeconds)]
								.filter((part) => part !== '')
								.join(' • ')}
						</span>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</section>
