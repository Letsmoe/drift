<script lang="ts">
	import { panels } from '#lib/layout/panels.svelte.ts';
	import { lyricsView } from '#lib/lyrics/lyricsView.svelte.ts';
	import type { LyricLine } from '#lib/api/lyrics.ts';
	import { player } from '#lib/player/player.svelte.ts';

	// The active line rests this far down the window, so the previous line stays visible.
	const ACTIVE_LINE_POSITION = 1 / 4;

	const activeIndex = $derived(lyricsView.activeLineIndex(player.currentTime));
	const lines = $derived(shownLines());
	let windowElement = $state<HTMLElement | null>(null);

	$effect(() => {
		if (player.track) {
			lyricsView.load(player.track);
		}
	});

	// Plain lyrics (activeIndex -1) stay at the top; synced ones follow playback.
	$effect(() => {
		if (!windowElement) {
			return;
		}
		const line = windowElement.querySelector<HTMLElement>(`[data-line-index="${activeIndex}"]`);
		if (!line) {
			windowElement.scrollTo({ top: 0 });
			return;
		}
		const targetTop = line.offsetTop - windowElement.clientHeight * ACTIVE_LINE_POSITION;
		windowElement.scrollTo({ top: Math.max(0, targetTop), behavior: 'smooth' });
	});

	function shownLines(): LyricLine[] {
		if (!lyricsView.lyrics) {
			return [];
		}
		return lyricsView.lyrics.lines;
	}
</script>

{#if lines.length > 0}
	<section class="flex flex-col gap-2 rounded-xl bg-raised/60 p-3">
		<div class="flex items-center justify-between">
			<h3 class="text-sm font-semibold text-muted">Lyrics</h3>
			<button
				type="button"
				onclick={() => panels.toggleFullPlayer()}
				class="rounded-full px-2 py-0.5 text-xs font-semibold text-dim hover:bg-hover hover:text-default"
			>
				Show all
			</button>
		</div>
		<div
			bind:this={windowElement}
			class="lyrics-window relative flex h-24 flex-col gap-0.5 overflow-hidden text-sm font-semibold"
		>
			{#each lines as line, index (index)}
				<p
					data-line-index={index}
					class={[
						'lyric-line min-h-5 truncate',
						index === activeIndex && 'text-default',
						index !== activeIndex && 'text-dim'
					]}
				>
					{line.text}
				</p>
			{/each}
		</div>
	</section>
{/if}

<style>
	/* Lines fade out at the bottom edge of the window. */
	.lyrics-window {
		mask-image: linear-gradient(to bottom, black 70%, transparent);
	}

	.lyric-line {
		transition: color var(--dur-fast);
	}

	@media (prefers-reduced-motion: reduce) {
		.lyric-line {
			transition: none;
		}
	}
</style>
