<script lang="ts">
	import { LoadingSpinner } from '@neoworks-dev/ui';
	import { lyricsView } from '#lib/lyrics/lyricsView.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';

	// After the user scrolls, auto-scroll stays off for this long.
	const AUTO_SCROLL_PAUSE_MS = 4000;
	// The active line rests this far from the top of the scroll area.
	const ACTIVE_LINE_POSITION = 1 / 3;

	const activeIndex = $derived(lyricsView.activeLineIndex(player.currentTime));
	const synced = $derived(lyricsView.lyrics?.synced === true);
	let activeLine = $state<HTMLElement | null>(null);
	let autoScrollResumesAt = 0;

	// Synced lyrics follow the playback position; plain lyrics have no active line.
	$effect(() => {
		if (activeLine && Date.now() >= autoScrollResumesAt) {
			scrollToActiveLine(activeLine);
		}
	});

	function scrollToActiveLine(line: HTMLElement): void {
		const scroller = findScrollParent(line);
		if (!scroller) {
			return;
		}
		const lineTop = line.getBoundingClientRect().top - scroller.getBoundingClientRect().top;
		const targetTop = scroller.clientHeight * ACTIVE_LINE_POSITION;
		scroller.scrollBy({ top: lineTop - targetTop, behavior: 'smooth' });
	}

	function findScrollParent(element: HTMLElement): HTMLElement | undefined {
		let parent = element.parentElement;
		while (parent) {
			if (parent.scrollHeight > parent.clientHeight && isScrollable(parent)) {
				return parent;
			}
			parent = parent.parentElement;
		}
		return undefined;
	}

	function isScrollable(element: HTMLElement): boolean {
		const overflowY = getComputedStyle(element).overflowY;
		return overflowY === 'auto' || overflowY === 'scroll';
	}

	function pauseAutoScroll(): void {
		autoScrollResumesAt = Date.now() + AUTO_SCROLL_PAUSE_MS;
	}

	function seekToLine(startMs: number): void {
		if (!synced) {
			return;
		}
		player.seek(startMs / 1000);
	}

	function lineClass(index: number): string {
		if (!synced) {
			return 'text-default';
		}
		if (index === activeIndex) {
			return 'text-default';
		}
		return 'text-default/35 hover:text-default/60';
	}
</script>

{#snippet lyricLine(text: string, startMs: number, index: number)}
	<button
		type="button"
		bind:this={
			() => (index === activeIndex ? activeLine : undefined),
			(element) => {
				if (index === activeIndex) {
					activeLine = element;
				}
			}
		}
		disabled={!synced}
		onclick={() => seekToLine(startMs)}
		class={[
			'lyric-line block w-full text-left text-3xl leading-snug font-bold enabled:cursor-pointer xl:text-4xl',
			lineClass(index)
		]}
	>
		{text}
	</button>
{/snippet}

<!-- Wheel, touch and drag scrolling bubble up here; they pause the auto-scroll. -->
<div
	role="presentation"
	onwheel={pauseAutoScroll}
	ontouchmove={pauseAutoScroll}
	onpointerdown={pauseAutoScroll}
	class="h-full"
>
	{#if lyricsView.status === 'loading'}
		<div class="flex h-full items-center justify-center"><LoadingSpinner size={28} /></div>
	{:else if lyricsView.status === 'ready' && lyricsView.lyrics}
		<div class="flex flex-col gap-3 py-[20vh] pr-6">
			{#each lyricsView.lyrics.lines as line, index (index)}
				{#if line.text === ''}
					<div class="h-6"></div>
				{:else}
					{@render lyricLine(line.text, line.startMs, index)}
				{/if}
			{/each}
			{#if lyricsView.lyrics.source}
				<p class="pt-8 text-sm text-dim">Source: {lyricsView.lyrics.source}</p>
			{/if}
		</div>
	{:else if lyricsView.status === 'failed'}
		<p class="flex h-full items-center justify-center text-xl font-semibold text-dim">
			Lyrics could not be loaded
		</p>
	{:else}
		<p class="flex h-full items-center justify-center text-xl font-semibold text-dim">
			No lyrics for this song
		</p>
	{/if}
</div>

<style>
	.lyric-line {
		transition: color var(--dur-fast);
	}

	@media (prefers-reduced-motion: reduce) {
		.lyric-line {
			transition: none;
		}
	}
</style>
