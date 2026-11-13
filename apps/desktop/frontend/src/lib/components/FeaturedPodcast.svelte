<script lang="ts">
	import { Button, View } from '@neoworks/svelte-gpui/ui';
	import { podcast } from '#lib/api/catalog.ts';
	import type { Item, PodcastPage } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import Label from '#lib/components/Label.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { formatDuration, itemHref } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { goto } from '#lib/router.svelte.ts';
	import { colors, radii } from '#lib/theme.ts';

	const latestEpisodeCount = 3;
	const coverSide = 208;

	let { item }: { item: Item } = $props();

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

	function openShow(): void {
		goto(itemHref('podcast', item.id));
	}
</script>

<View
	style={{
		direction: 'row',
		gap: 24,
		padding: 20,
		radius: radii.extraLarge,
		background: colors.raised
	}}
>
	<Button style={{ shrink: 0, alignSelf: 'start' }} onpress={openShow}>
		<Cover
			thumbnails={item.thumbnails}
			size={coverSide}
			style={{
				size: coverSide,
				radius: radii.large,
				shadow: { y: 12, blur: 32, color: 'rgba(0, 0, 0, 0.55)' }
			}}
		/>
	</Button>
	<View style={{ flex: 1, minWidth: 0, gap: 8 }}>
		<Label size="tiny" tone="dim" weight="semibold" style={{ uppercase: true }}>
			Featured podcast
		</Label>
		<View
			style={{ alignSelf: 'start', minWidth: 0, cursor: 'pointer' }}
			hover={{ underline: true }}
			onpress={openShow}
		>
			<Label size="title" weight="bold" lines={2}>{item.title}</Label>
		</View>
		<Label size="small" tone="muted">{item.subtitle}</Label>
		{#if item.description}
			<Label size="small" tone="dim" lines={3} style={{ maxWidth: 768 }}>{item.description}</Label>
		{/if}
		{#if episodes.length > 0}
			<View style={{ marginTop: 8 }}>
				<Label size="tiny" tone="muted" weight="semibold" style={{ paddingBottom: 4 }}>
					Latest episodes
				</Label>
				{#each episodes as episode, index (episode.track.videoId)}
					<View
						style={{
							direction: 'row',
							align: 'center',
							gap: 12,
							paddingX: 8,
							paddingY: 6,
							radius: radii.medium
						}}
						hover={{ background: colors.hover }}
					>
						<PlayOverlayButton
							size="sm"
							playing={player.isPlaying(episode.track)}
							onpress={() => playEpisode(index)}
						/>
						<Label
							size="small"
							accented={player.isCurrent(episode.track)}
							truncate
							style={{ flex: 1, minWidth: 0 }}
						>
							{episode.track.title}
						</Label>
						<Label size="tiny" tone="dim" style={{ shrink: 0 }}>
							{[episode.published, formatDuration(episode.track.durationSeconds)]
								.filter((part) => part !== '')
								.join(' • ')}
						</Label>
					</View>
				{/each}
			</View>
		{/if}
	</View>
</View>
