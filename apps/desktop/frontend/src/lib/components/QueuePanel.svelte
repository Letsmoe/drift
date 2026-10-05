<script lang="ts">
	import { FloatingScrollbar } from '@neoworks-dev/ui';
	import BroomIcon from 'phosphor-svelte/lib/BroomIcon';
	import SparkleIcon from 'phosphor-svelte/lib/SparkleIcon';
	import Cover from '#lib/components/Cover.svelte';
	import QueueRow from '#lib/components/QueueRow.svelte';
	import TrackCredits from '#lib/components/TrackCredits.svelte';
	import { player } from '#lib/player/player.svelte.ts';
	import { contextLabel } from '#lib/queue/contextLabel.ts';
	import { splitQueue, type UpcomingEntry } from '#lib/queue/queueSections.ts';

	const sections = $derived(splitQueue(player.queue));
	const hasUpcoming = $derived(sections.queued.length > 0 || sections.autoplay.length > 0);
	const queuedHeading = $derived(buildQueuedHeading(player.queue.context));

	let draggedEntryId = $state<number | undefined>(undefined);
	let dropTargetEntryId = $state<number | undefined>(undefined);

	function buildQueuedHeading(context: string): string {
		const label = contextLabel(context);
		if (label === '') {
			return 'Next in queue';
		}
		return `Next from: ${label}`;
	}

	function autoplayButtonLabel(): string {
		if (player.queue.autoplay) {
			return 'Autoplay on';
		}
		return 'Autoplay off';
	}

	function startDrag(event: DragEvent, entryId: number): void {
		draggedEntryId = entryId;
		event.dataTransfer?.setData('text/plain', String(entryId));
	}

	function dragOver(event: DragEvent, entryId: number): void {
		if (draggedEntryId === undefined) {
			return;
		}
		event.preventDefault();
		dropTargetEntryId = entryId;
	}

	function drop(event: DragEvent, target: UpcomingEntry): void {
		event.preventDefault();
		const movedEntryId = draggedEntryId;
		endDrag();
		if (movedEntryId === undefined || movedEntryId === target.entry.id) {
			return;
		}
		player.move(movedEntryId, target.index);
	}

	function endDrag(): void {
		draggedEntryId = undefined;
		dropTargetEntryId = undefined;
	}
</script>

{#snippet upcomingList(entries: UpcomingEntry[])}
	<div role="list" class="flex flex-col">
		{#each entries as upcoming (upcoming.entry.id)}
			<QueueRow
				entry={upcoming.entry}
				draggable
				dropTarget={dropTargetEntryId === upcoming.entry.id}
				ondragstart={(event) => startDrag(event, upcoming.entry.id)}
				ondragover={(event) => dragOver(event, upcoming.entry.id)}
				ondrop={(event) => drop(event, upcoming)}
				ondragend={endDrag}
			/>
		{/each}
	</div>
{/snippet}

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
					<Cover
						thumbnails={sections.nowPlaying.track.thumbnails}
						size={288}
						class="w-full rounded-xl shadow-lg"
					/>
					<div class="min-w-0 [&>p:first-child]:text-lg [&>p:first-child]:font-semibold">
						<TrackCredits track={sections.nowPlaying.track} />
					</div>
				</section>
			{:else}
				<p class="px-2 py-8 text-center text-sm text-dim">
					Nothing is queued. Play a song to start.
				</p>
			{/if}

			{#if sections.queued.length > 0}
				<section class="flex flex-col gap-1">
					<h3 class="truncate px-1.5 text-sm font-semibold text-muted">{queuedHeading}</h3>
					{@render upcomingList(sections.queued)}
				</section>
			{/if}

			{#if sections.autoplay.length > 0}
				<section class="flex flex-col gap-1">
					<h3 class="px-1.5 text-sm font-semibold text-muted">Autoplay: Recommended for you</h3>
					{@render upcomingList(sections.autoplay)}
				</section>
			{/if}

			{#if sections.nowPlaying && !hasUpcoming}
				<p class="px-2 text-center text-sm text-dim">No songs up next.</p>
			{/if}
		</div>
	</FloatingScrollbar>
</aside>
