<script lang="ts">
	import { Checkbox, LoadingSpinner } from '@neoworks-dev/ui';
	import { lyricsSettings } from '#lib/lyrics/lyricsSettings.svelte.ts';
	import { lyricsView } from '#lib/lyrics/lyricsView.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';

	// After the user scrolls, auto-scroll stays off for this long.
	const AUTO_SCROLL_PAUSE_MS = 4000;
	// The active line rests this far from the top of the scroll area.
	const ACTIVE_LINE_POSITION = 1 / 3;

	const activeIndex = $derived(lyricsView.activeLineIndex(player.currentTime));
	const synced = $derived(lyricsView.lyrics?.synced === true);
	let linesElement = $state<HTMLElement | null>(null);
	let autoScrollResumesAt = 0;

	// Synced lyrics follow the playback position; plain lyrics have no active line.
	$effect(() => {
		const line = findLine(linesElement, activeIndex);
		if (line && Date.now() >= autoScrollResumesAt) {
			scrollToActiveLine(line);
		}
	});

	function findLine(container: HTMLElement | null, index: number): HTMLElement | null {
		if (!container || index < 0) {
			return null;
		}
		return container.querySelector<HTMLElement>(`[data-line-index="${index}"]`);
	}

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
		// A click is a pointerdown too; picking a line means following it again.
		autoScrollResumesAt = 0;
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

	// LRCLIB only matters when YouTube has no synced lyrics, or when it supplied them.
	function offersLRCLIB(): boolean {
		const shown = lyricsView.lyrics;
		if (!shown) {
			return true;
		}
		return !shown.synced || shown.source === 'LRCLIB';
	}
</script>

{#snippet lyricLine(text: string, startMs: number, index: number)}
	<button
		type="button"
		data-line-index={index}
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

{#snippet lrclibToggle()}
	<label class="flex max-w-md cursor-pointer items-start gap-3 text-sm text-dim">
		<span class="pt-0.5">
			<Checkbox
				bind:checked={
					() => lyricsSettings.allowLRCLIB, (allow) => lyricsSettings.setAllowLRCLIB(allow)
				}
			/>
		</span>
		<span>
			<span class="block font-medium text-muted">Find synced lyrics on LRCLIB</span>
			When YouTube has no timed lyrics, the song's title, artist, album and length are sent to lrclib.net.
		</span>
	</label>
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
		<div bind:this={linesElement} class="flex flex-col gap-3 py-[20vh] pr-6">
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
			{#if offersLRCLIB()}
				<div class="pt-4">{@render lrclibToggle()}</div>
			{/if}
		</div>
	{:else if lyricsView.status === 'failed'}
		<div class="flex h-full flex-col items-center justify-center gap-6">
			<p class="text-xl font-semibold text-dim">Lyrics could not be loaded</p>
			{@render lrclibToggle()}
		</div>
	{:else}
		<div class="flex h-full flex-col items-center justify-center gap-6">
			<p class="text-xl font-semibold text-dim">No lyrics for this song</p>
			{@render lrclibToggle()}
		</div>
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
