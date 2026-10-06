<script lang="ts">
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
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
		draggable = false,
		dropTarget = false,
		ondragstart,
		ondragover,
		ondrop,
		ondragend
	}: {
		entry: QueueEntry;
		current?: boolean;
		draggable?: boolean;
		dropTarget?: boolean;
		ondragstart?: (event: DragEvent) => void;
		ondragover?: (event: DragEvent) => void;
		ondrop?: (event: DragEvent) => void;
		ondragend?: (event: DragEvent) => void;
	} = $props();

	const contextMenu = new PopupMenu();

	$effect(() => {
		downloads.ensureStates([entry.track.videoId]);
	});
</script>

<div
	role="listitem"
	{draggable}
	{ondragstart}
	{ondragover}
	{ondrop}
	{ondragend}
	oncontextmenu={(event) => contextMenu.openAtPointer(event)}
	class={[
		'group/queuerow duration-fast flex items-center gap-2.5 rounded-md px-1.5 py-1 transition-colors hover:bg-hover',
		dropTarget && 'border-t-2 border-action'
	]}
>
	<Cover thumbnails={entry.track.thumbnails} size={32} class="size-8 shrink-0 rounded" />
	<div class="min-w-0 flex-1">
		<div class="flex min-w-0 items-center gap-1.5">
			<p class={['truncate text-[13px] leading-tight font-medium', current && 'text-action']}>
				{entry.track.title}
			</p>
			<DownloadedIndicator videoId={entry.track.videoId} size={12} />
		</div>
		<p class="truncate text-xs leading-tight text-dim">{formatArtists(entry.track.artists)}</p>
	</div>
	{#if !current}
		<div class="flex shrink-0 items-center opacity-0 group-hover/queuerow:opacity-100">
			<button
				type="button"
				aria-label="Play {entry.track.title}"
				class="rounded-full p-1 text-muted hover:text-default"
				onclick={() => player.jump(entry.id)}
			>
				<PlayIcon size={14} weight="fill" />
			</button>
			<button
				type="button"
				aria-label="Remove {entry.track.title} from queue"
				class="rounded-full p-1 text-muted hover:text-default"
				onclick={() => player.remove(entry.id)}
			>
				<XIcon size={14} />
			</button>
		</div>
	{/if}
</div>

<TrackContextMenu popup={contextMenu} track={entry.track} kind={entry.kind as ItemKind} />
