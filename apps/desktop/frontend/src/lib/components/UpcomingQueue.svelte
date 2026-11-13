<script lang="ts">
	import { Button, DropTarget, Draggable, View } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Label from '#lib/components/Label.svelte';
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

	function showAllLabel(): string {
		if (showAllPlayed) {
			return 'Show less';
		}
		return `Show all ${sections.played.length}`;
	}

	// What a dragged row carries is its entry's ID.
	function isEntryId(data: unknown): boolean {
		return typeof data === 'number';
	}

	function drop(movedEntryId: unknown, target: UpcomingEntry): void {
		if (typeof movedEntryId !== 'number' || movedEntryId === target.entry.id) {
			return;
		}
		player.move(movedEntryId, target.index);
	}
</script>

{#snippet upcomingList(entries: UpcomingEntry[])}
	<View>
		{#each entries as upcoming (upcoming.entry.id)}
			<DropTarget
				accept={isEntryId}
				ondrop={({ data }) => drop(data, upcoming)}
				over={{ borderTopWidth: 2, borderColor: appearance.accent }}
			>
				<Draggable data={upcoming.entry.id}>
					<QueueRow entry={upcoming.entry} {size} />
				</Draggable>
			</DropTarget>
		{/each}
	</View>
{/snippet}

{#snippet heading(text: string)}
	<View style={{ paddingX: 6 }}>
		<Label size="small" weight="semibold" tone="muted" truncate>{text}</Label>
	</View>
{/snippet}

{#if sections.played.length > 0}
	<View style={{ gap: 4 }}>
		<View style={{ direction: 'row', align: 'center', justify: 'between', gap: 8, paddingX: 6 }}>
			<Label size="small" weight="semibold" tone="muted">Played</Label>
			{#if sections.played.length > collapsedPlayedCount}
				<Button hover={{ color: 'white' }} onpress={() => (showAllPlayed = !showAllPlayed)}>
					<Label size="tiny" weight="semibold" tone="dim">{showAllLabel()}</Label>
				</Button>
			{/if}
		</View>
		<View>
			{#each visiblePlayed as playedEntry (playedEntry.entry.id)}
				<QueueRow entry={playedEntry.entry} {size} played />
			{/each}
		</View>
	</View>
{/if}

{#if sections.queued.length > 0}
	<View style={{ gap: 4 }}>
		{@render heading(queuedHeading)}
		{@render upcomingList(sections.queued)}
	</View>
{/if}

{#if sections.autoplay.length > 0}
	<View style={{ gap: 4 }}>
		{@render heading('Autoplay: Recommended for you')}
		{@render upcomingList(sections.autoplay)}
	</View>
{/if}

{#if sections.nowPlaying && !hasUpcoming}
	<View style={{ paddingX: 8, align: 'center' }}>
		<Label size="small" tone="dim">No songs up next.</Label>
	</View>
{/if}
