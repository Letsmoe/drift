<script lang="ts">
	import { panels } from '#lib/layout/panels.svelte.ts';
	import { lyricsView } from '#lib/lyrics/lyricsView.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';

	// Lines shown: the active one and those after it (synced), or the first ones (plain).
	const previewLineCount = 4;

	const activeIndex = $derived(lyricsView.activeLineIndex(player.currentTime));
	const preview = $derived(previewLines());

	$effect(() => {
		if (player.track) {
			lyricsView.load(player.track.videoId);
		}
	});

	function previewLines(): { text: string; active: boolean }[] {
		const lyrics = lyricsView.lyrics;
		if (!lyrics || lyrics.lines.length === 0) {
			return [];
		}
		const start = Math.max(0, activeIndex);
		return lyrics.lines.slice(start, start + previewLineCount).map((line, offset) => ({
			text: line.text,
			active: start + offset === activeIndex
		}));
	}
</script>

{#if preview.length > 0}
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
		<div class="flex flex-col gap-0.5 text-sm font-semibold">
			{#each preview as line, index (index)}
				<p class={['truncate', line.active && 'text-default', !line.active && 'text-dim']}>
					{line.text}
				</p>
			{/each}
		</div>
	</section>
{/if}
