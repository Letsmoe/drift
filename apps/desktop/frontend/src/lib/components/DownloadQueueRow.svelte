<script lang="ts">
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import type { DownloadItem } from '#lib/api/downloads.ts';
	import Cover from '#lib/components/Cover.svelte';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { downloadFraction, formatBytes } from '#lib/downloads/downloadProgress.ts';
	import { formatArtists } from '#lib/format.ts';

	let { item }: { item: DownloadItem } = $props();

	const videoId = $derived(item.track.videoId);
	// The live status from events is fresher than the item loaded with the list.
	const status = $derived(currentStatus());
	const failed = $derived(status.state === 'failed');
	const percent = $derived(Math.round(downloadFraction(status) * 100));

	function currentStatus() {
		const live = downloads.statusOf(videoId);
		if (live) {
			return live;
		}
		return item.status;
	}

	function describeState(): string {
		switch (status.state) {
			case 'downloading':
				return `${formatBytes(status.bytesDone)} of ${formatBytes(status.bytesTotal)}`;
			case 'paused':
				return 'Paused';
			case 'failed':
				return status.error;
			default:
				return 'Waiting';
		}
	}
</script>

<div class="group/queuerow flex items-center gap-3 rounded-md px-4 py-2 hover:bg-hover">
	<Cover thumbnails={item.track.thumbnails} size={40} class="size-10 shrink-0 rounded-sm" />
	<div class="min-w-0 flex-1">
		<p class="truncate text-base">{item.track.title}</p>
		<p class="truncate text-sm text-dim">{formatArtists(item.track.artists)}</p>
	</div>
	<div class="flex w-64 flex-col gap-1">
		<p class={['truncate text-right text-xs', failed && 'text-red', !failed && 'text-dim']}>
			{describeState()}
		</p>
		{#if status.state === 'downloading'}
			<div class="h-1 overflow-hidden rounded-full bg-raised">
				<div class="h-full bg-action" style:width="{percent}%"></div>
			</div>
		{/if}
	</div>
	{#if failed}
		<button
			type="button"
			aria-label="Retry download of {item.track.title}"
			onclick={() => downloads.retry(videoId)}
			class="text-dim hover:text-default"
		>
			<ArrowClockwiseIcon size={18} />
		</button>
	{/if}
	<button
		type="button"
		aria-label="Cancel download of {item.track.title}"
		onclick={() => downloads.cancel(videoId)}
		class="text-dim hover:text-default"
	>
		<XIcon size={18} />
	</button>
</div>
