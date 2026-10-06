<script lang="ts">
	import QueueRow from '#lib/components/QueueRow.svelte';
	import { player } from '#lib/player/player.svelte.ts';
	import { contextLabel } from '#lib/queue/contextLabel.ts';
	import { splitQueue, type UpcomingEntry } from '#lib/queue/queueSections.ts';

	// Played history, then the queued and autoplay sections with drag-to-reorder;
	// shared by the Now playing column and the full-screen player.
	let { size = 'compact' }: { size?: 'compact' | 'comfortable' } = $props();

	// Played entries shown before "Show all".
	const collapsedPlayedCount = 3;

	const sections = $derived(splitQueue(player.queue));
	const hasUpcoming = $derived(sections.queued.length > 0 || sections.autoplay.length > 0);
	const queuedHeading = $derived(buildQueuedHeading(player.queue.context));

	let showAllPlayed = $state(false);
	const visiblePlayed = $derived(visiblePlayedEntries());

	let draggedEntryId = $state<number | undefined>(undefined);
	let dropTargetEntryId = $state<number | undefined>(undefined);

	function buildQueuedHeading(context: string): string {
		const label = contextLabel(context);
		let heading = `Next from: ${label}`;
		if (label === '') {
			heading = 'Next in queue';
		}
		if (player.queue.magicShuffle) {
			return `${heading} · magic shuffle`;
		}
		return heading;
	}

	function visiblePlayedEntries(): UpcomingEntry[] {
		if (showAllPlayed) {
			return sections.played;
		}
		return sections.played.slice(-collapsedPlayedCount);
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
				{size}
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

{#if sections.played.length > 0}
	<section class="flex flex-col gap-1">
		<div class="flex items-center justify-between gap-2 px-1.5">
			<h3 class="text-sm font-semibold text-muted">Played</h3>
			{#if sections.played.length > collapsedPlayedCount}
				<button
					type="button"
					class="text-xs font-semibold text-dim hover:text-default"
					onclick={() => (showAllPlayed = !showAllPlayed)}
				>
					{#if showAllPlayed}Show less{:else}Show all {sections.played.length}{/if}
				</button>
			{/if}
		</div>
		<div role="list" class="flex flex-col">
			{#each visiblePlayed as playedEntry (playedEntry.entry.id)}
				<QueueRow entry={playedEntry.entry} {size} played />
			{/each}
		</div>
	</section>
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
