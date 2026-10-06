<script lang="ts">
	import ArrowCircleDownIcon from 'phosphor-svelte/lib/ArrowCircleDownIcon';
	import ClockIcon from 'phosphor-svelte/lib/ClockIcon';
	import PauseCircleIcon from 'phosphor-svelte/lib/PauseCircleIcon';
	import WarningCircleIcon from 'phosphor-svelte/lib/WarningCircleIcon';
	import type { ItemKind, Track } from '#lib/api/music.ts';
	import ProgressRing from '#lib/components/ProgressRing.svelte';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { downloadFraction } from '#lib/downloads/downloadProgress.ts';

	let {
		track,
		kind = 'song',
		size = 18,
		revealOnHover = true
	}: { track: Track; kind?: ItemKind; size?: number; revealOnHover?: boolean } = $props();

	const status = $derived(downloads.statusOf(track.videoId));
	const state = $derived(downloads.stateOf(track.videoId));
	const label = $derived(describeAction(state));
	// In rows, idle and finished downloads only show on hover; work in progress and failures always show.
	const alwaysVisible = $derived(!revealOnHover || (state !== '' && state !== 'done'));

	function describeAction(current: string): string {
		switch (current) {
			case 'queued':
			case 'downloading':
				return `Cancel download of ${track.title}`;
			case 'paused':
				return `Cancel paused download of ${track.title}`;
			case 'failed':
				return `Retry download of ${track.title}`;
			case 'done':
				return `Remove download of ${track.title}`;
			default:
				return `Download ${track.title}`;
		}
	}

	function act(event: MouseEvent) {
		event.stopPropagation();
		switch (state) {
			case '':
				downloads.download(track, kind);
				return;
			case 'failed':
				downloads.retry(track.videoId);
				return;
			case 'done':
				downloads.remove(track.videoId);
				return;
			default:
				downloads.cancel(track.videoId);
		}
	}
</script>

<button
	type="button"
	aria-label={label}
	title={label}
	onclick={act}
	class={[
		'flex items-center justify-center hover:text-default focus-visible:opacity-100',
		'group-hover/row:opacity-100',
		!alwaysVisible && 'opacity-0',
		state === 'done' && 'text-action hover:text-action',
		state === 'failed' && 'text-red hover:text-red'
	]}
>
	{#if state === 'downloading'}
		<ProgressRing fraction={downloadFraction(status)} {size} />
	{:else if state === 'queued'}
		<ClockIcon {size} />
	{:else if state === 'paused'}
		<PauseCircleIcon {size} />
	{:else if state === 'failed'}
		<WarningCircleIcon {size} weight="fill" />
	{:else if state === 'done'}
		<ArrowCircleDownIcon {size} weight="fill" />
	{:else}
		<ArrowCircleDownIcon {size} />
	{/if}
</button>
