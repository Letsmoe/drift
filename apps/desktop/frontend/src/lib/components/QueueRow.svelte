<script lang="ts">
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import type { QueueEntry } from '#lib/api/queue.ts';
	import Cover from '#lib/components/Cover.svelte';
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
</script>

<div
	role="listitem"
	{draggable}
	{ondragstart}
	{ondragover}
	{ondrop}
	{ondragend}
	class={[
		'group/queuerow duration-fast flex items-center gap-3 rounded-md p-1.5 transition-colors hover:bg-hover',
		dropTarget && 'border-t-2 border-green'
	]}
>
	<Cover thumbnails={entry.track.thumbnails} size={40} class="size-10 shrink-0 rounded" />
	<div class="min-w-0 flex-1">
		<p class={['truncate text-sm font-medium', current && 'text-green']}>{entry.track.title}</p>
		<p class="truncate text-xs text-dim">{formatArtists(entry.track.artists)}</p>
	</div>
	{#if !current}
		<div class="flex shrink-0 items-center opacity-0 group-hover/queuerow:opacity-100">
			<button
				type="button"
				aria-label="Play {entry.track.title}"
				class="rounded-full p-1.5 text-muted hover:text-default"
				onclick={() => player.jump(entry.id)}
			>
				<PlayIcon size={16} weight="fill" />
			</button>
			<button
				type="button"
				aria-label="Remove {entry.track.title} from queue"
				class="rounded-full p-1.5 text-muted hover:text-default"
				onclick={() => player.remove(entry.id)}
			>
				<XIcon size={16} />
			</button>
		</div>
	{/if}
</div>
