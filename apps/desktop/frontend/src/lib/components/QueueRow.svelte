<script lang="ts">
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import SparkleIcon from 'phosphor-svelte/lib/SparkleIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import type { ItemKind } from '#lib/api/music.ts';
	import type { QueueEntry } from '#lib/api/queue.ts';
	import Cover from '#lib/components/Cover.svelte';
	import DownloadedIndicator from '#lib/components/DownloadedIndicator.svelte';
	import TrackContextMenu from '#lib/components/TrackContextMenu.svelte';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { formatArtists } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';

	let {
		entry,
		current = false,
		played = false,
		draggable = false,
		dropTarget = false,
		size = 'compact',
		ondragstart,
		ondragover,
		ondrop,
		ondragend
	}: {
		entry: QueueEntry;
		current?: boolean;
		/** Already played: shown dimmed, above the current entry. */
		played?: boolean;
		draggable?: boolean;
		dropTarget?: boolean;
		/** Compact rows fit the side panel; comfortable ones the full-screen player. */
		size?: 'compact' | 'comfortable';
		ondragstart?: (event: DragEvent) => void;
		ondragover?: (event: DragEvent) => void;
		ondrop?: (event: DragEvent) => void;
		ondragend?: (event: DragEvent) => void;
	} = $props();

	const contextMenu = new PopupMenu();

	$effect(() => {
		downloads.ensureStates([entry.track.videoId]);
	});

	function playEntry(): void {
		if (!current) {
			player.jump(entry.id);
		}
	}

	function removeEntry(event: MouseEvent): void {
		event.stopPropagation();
		player.remove(entry.id);
	}
</script>

<!-- Clicking anywhere on the row plays it; keyboard users have the row's play button. -->
<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
<div
	role="listitem"
	{draggable}
	{ondragstart}
	{ondragover}
	{ondrop}
	{ondragend}
	oncontextmenu={(event) => contextMenu.openAtPointer(event)}
	onclick={playEntry}
	class={[
		'group/queuerow duration-fast flex items-center rounded-md transition-colors hover:bg-hover',
		size === 'compact' && 'gap-2.5 px-1.5 py-1',
		size === 'comfortable' && 'gap-3 p-2',
		dropTarget && 'border-t-2 border-action',
		played && 'opacity-60 hover:opacity-100',
		!current && 'cursor-pointer'
	]}
>
	{#if size === 'comfortable'}
		<Cover thumbnails={entry.track.thumbnails} size={48} class="size-12 shrink-0 rounded-md" />
	{:else}
		<Cover thumbnails={entry.track.thumbnails} size={32} class="size-8 shrink-0 rounded" />
	{/if}
	<div class="min-w-0 flex-1">
		<div class="flex min-w-0 items-center gap-1.5">
			<p
				class={[
					'truncate font-medium',
					size === 'compact' && 'text-[13px] leading-tight',
					size === 'comfortable' && 'text-base',
					current && 'text-action'
				]}
			>
				{entry.track.title}
			</p>
			{#if entry.magic}
				<span title="Recommended by magic shuffle" class="shrink-0 text-action">
					<SparkleIcon size={12} weight="fill" />
				</span>
			{/if}
			<DownloadedIndicator videoId={entry.track.videoId} size={12} />
		</div>
		<p
			class={[
				'truncate text-dim',
				size === 'compact' && 'text-xs leading-tight',
				size === 'comfortable' && 'text-sm'
			]}
		>
			{formatArtists(entry.track.artists)}
		</p>
	</div>
	{#if !current}
		<div class="flex shrink-0 items-center opacity-0 group-hover/queuerow:opacity-100">
			<button
				type="button"
				aria-label="Play {entry.track.title}"
				class="rounded-full p-1 text-muted hover:text-default"
				onclick={(event) => {
					event.stopPropagation();
					playEntry();
				}}
			>
				<PlayIcon size={14} weight="fill" />
			</button>
			<button
				type="button"
				aria-label="Remove {entry.track.title} from queue"
				class="rounded-full p-1 text-muted hover:text-default"
				onclick={removeEntry}
			>
				<XIcon size={14} />
			</button>
		</div>
	{/if}
</div>

<TrackContextMenu popup={contextMenu} track={entry.track} kind={entry.kind as ItemKind} />
