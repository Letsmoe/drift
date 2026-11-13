<!--
  A window of the playing track's lyrics. Synced lyrics keep the active line a quarter of the way down so the
  previous line stays visible; plain lyrics stay at the top. The old window faded out at the bottom with a CSS
  mask, which `Style` cannot do.
-->
<script lang="ts">
	import { Button, Scroll, View } from '@neoworks/svelte-gpui/ui';
	import type { LyricLine } from '#lib/api/lyrics.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Label from '#lib/components/Label.svelte';
	import { panels } from '#lib/layout/panels.svelte.ts';
	import { lyricsView } from '#lib/lyrics/lyricsView.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	const windowHeight = 144;
	const lineHeight = 28;
	const lineGap = 4;
	// The active line rests this far down the window.
	const activeLinePosition = 1 / 4;

	const activeIndex = $derived(lyricsView.activeLineIndex(player.currentTime));
	const lines = $derived(shownLines());
	let scroller: ReturnType<typeof Scroll> | undefined = $state();

	$effect(() => {
		if (player.track) {
			lyricsView.load(player.track);
		}
	});

	// Plain lyrics (activeIndex -1) stay at the top; synced ones follow playback.
	$effect(() => {
		if (activeIndex < 0) {
			scroller?.scrollTo(0);
			return;
		}
		const lineTop = activeIndex * (lineHeight + lineGap);
		scroller?.scrollTo(Math.max(0, lineTop - windowHeight * activeLinePosition));
	});

	function lineColor(index: number): string {
		if (index === activeIndex) {
			return colors.text;
		}
		return appearance.dim;
	}

	function shownLines(): LyricLine[] {
		if (!lyricsView.lyrics) {
			return [];
		}
		return lyricsView.lyrics.lines;
	}
</script>

{#if lines.length > 0}
	<View style={{ gap: 8 }}>
		<View style={{ direction: 'row', align: 'center', justify: 'between' }}>
			<Label size="small" weight="semibold" tone="muted">Lyrics</Label>
			<Button
				style={{
					paddingX: 8,
					paddingY: 2,
					radius: radii.full,
					color: appearance.dim,
					transition: hoverTransition
				}}
				hover={{ background: colors.hover, color: colors.text }}
				onpress={() => panels.toggleFullPlayer()}
			>
				<Label size="tiny" weight="semibold">Show all</Label>
			</Button>
		</View>
		<Scroll bind:this={scroller} style={{ height: windowHeight }}>
			<View style={{ gap: lineGap }}>
				{#each lines as line, index (index)}
					<Label
						size="large"
						weight="semibold"
						truncate
						style={{
							height: lineHeight,
							color: lineColor(index)
						}}
					>
						{line.text}
					</Label>
				{/each}
			</View>
		</Scroll>
	</View>
{/if}
