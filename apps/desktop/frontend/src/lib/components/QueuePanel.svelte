<script lang="ts">
	import { FloatingScrollbar } from '@neoworks-dev/ui';
	import BroomIcon from 'phosphor-svelte/lib/BroomIcon';
	import SparkleIcon from 'phosphor-svelte/lib/SparkleIcon';
	import Cover from '#lib/components/Cover.svelte';
	import TrackCredits from '#lib/components/TrackCredits.svelte';
	import UpcomingQueue from '#lib/components/UpcomingQueue.svelte';
	import { panels } from '#lib/layout/panels.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { splitQueue } from '#lib/queue/queueSections.ts';

	const sections = $derived(splitQueue(player.queue));

	function autoplayButtonLabel(): string {
		if (player.queue.autoplay) {
			return 'Autoplay on';
		}
		return 'Autoplay off';
	}
</script>

<aside class="art-wash-strong flex w-80 shrink-0 flex-col rounded-2xl">
	<header class="flex h-12 shrink-0 items-center justify-between gap-2 px-4">
		<h2 class="text-base font-semibold">Now playing</h2>
		<div class="flex items-center gap-1">
			<button
				type="button"
				aria-pressed={player.queue.autoplay}
				onclick={() => player.toggleAutoplay()}
				class={[
					'duration-fast flex items-center gap-1 rounded-full px-2 py-1 text-xs transition-colors hover:bg-hover',
					player.queue.autoplay && 'text-action',
					!player.queue.autoplay && 'text-dim'
				]}
			>
				<SparkleIcon size={14} />
				{autoplayButtonLabel()}
			</button>
			<button
				type="button"
				aria-label="Clear queue"
				title="Clear queue"
				disabled={player.queue.entries.length === 0}
				onclick={() => player.clearQueue()}
				class="rounded-full p-1.5 text-muted hover:bg-hover hover:text-default disabled:opacity-40"
			>
				<BroomIcon size={16} />
			</button>
		</div>
	</header>

	<FloatingScrollbar class="min-h-0 flex-1">
		<div class="flex flex-col gap-4 px-2 pb-4">
			{#if sections.nowPlaying}
				<section class="flex flex-col gap-3 px-2">
					<button
						type="button"
						aria-label="Open full-screen player"
						onclick={() => panels.toggleFullPlayer()}
						class="block w-full cursor-pointer rounded-xl"
					>
						<Cover
							thumbnails={sections.nowPlaying.track.thumbnails}
							size={288}
							class="w-full rounded-xl shadow-lg"
						/>
					</button>
					<div class="min-w-0 [&>p:first-child]:text-lg [&>p:first-child]:font-semibold">
						<TrackCredits track={sections.nowPlaying.track} />
					</div>
				</section>
			{:else}
				<p class="px-2 py-8 text-center text-sm text-dim">
					Nothing is queued. Play a song to start.
				</p>
			{/if}

			<UpcomingQueue />
		</div>
	</FloatingScrollbar>
</aside>
