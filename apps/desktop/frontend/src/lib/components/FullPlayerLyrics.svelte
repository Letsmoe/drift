<!--
  The lyrics of the playing song. Synced lyrics follow the position: the current line is bright and rests a
  third of the way down, and pressing a line seeks to it. Scrolling by hand pauses the following for a moment.
-->
<script lang="ts">
	import { window } from '@neoworks/svelte-gpui';
	import { Button, Scroll, View, type ScrollPosition } from '@neoworks/svelte-gpui/ui';
	import { Checkbox, LoadingSpinner } from '@neoworks/svelte-gpui-ui';
	import { cubicOut } from 'svelte/easing';
	import { Tween } from 'svelte/motion';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Label from '#lib/components/Label.svelte';
	import { lyricsSettings } from '#lib/lyrics/lyricsSettings.svelte.ts';
	import { lyricsView } from '#lib/lyrics/lyricsView.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { colors, hoverTransition } from '#lib/theme.ts';

	// After the user scrolls, following stays off for this long.
	const autoScrollPauseMilliseconds = 4000;
	// A scroll report this soon after one of ours is ours.
	const ownScrollWindowMilliseconds = 150;
	// The current line rests this far from the top of the scroll area.
	const activeLinePosition = 1 / 3;
	const lineGap = 12;
	const blankLineHeight = 24;
	const lineFontSize = 30;
	const lineHeight = 41;
	const inactiveLineColor = 'rgba(250, 250, 250, 0.35)';
	const hoveredLineColor = 'rgba(250, 250, 250, 0.6)';

	const activeIndex = $derived(lyricsView.activeLineIndex(player.currentTime));
	const synced = $derived(lyricsView.lyrics?.synced === true);
	const lines = $derived(lyricsView.lyrics?.lines ?? []);
	// Above the first line and below the last one, so every line can reach the resting place.
	const edgePadding = $derived(Math.round(window.size.height * 0.2));

	let scroller: Scroll | undefined = $state();
	let viewportHeight = $state(0);
	let lineHeights = $state<number[]>([]);
	const scrollOffset = new Tween(0, { duration: 450, easing: cubicOut });
	let autoScrollResumesAt = 0;
	let ownScrollAt = 0;
	let lastSeenOffset = 0;

	// A new set of lines starts at the top, with nothing measured.
	$effect(() => {
		lineHeights = new Array(lines.length).fill(0);
		autoScrollResumesAt = 0;
		void scrollOffset.set(0, { duration: 0 });
	});

	$effect(() => {
		const target = restingOffset(activeIndex);
		if (target === undefined || Date.now() < autoScrollResumesAt) {
			return;
		}
		void scrollOffset.set(target);
	});

	$effect(() => {
		ownScrollAt = performance.now();
		scroller?.scrollTo(scrollOffset.current);
	});

	function restingOffset(index: number): number | undefined {
		if (index < 0 || index >= lineHeights.length) {
			return undefined;
		}
		let lineTop = edgePadding;
		for (let before = 0; before < index; before += 1) {
			lineTop += lineHeights[before] + lineGap;
		}
		return Math.max(0, lineTop - viewportHeight * activeLinePosition);
	}

	function recordLineHeight(index: number, height: number): void {
		lineHeights[index] = height;
	}

	function handleScroll(position: ScrollPosition): void {
		const moved = Math.abs(position.y - lastSeenOffset) > 0.5;
		lastSeenOffset = position.y;
		if (!moved || performance.now() - ownScrollAt < ownScrollWindowMilliseconds) {
			return;
		}
		autoScrollResumesAt = Date.now() + autoScrollPauseMilliseconds;
		void scrollOffset.set(position.y, { duration: 0 });
	}

	function seekToLine(startMilliseconds: number): void {
		if (!synced) {
			return;
		}
		// Picking a line means following it again.
		autoScrollResumesAt = 0;
		player.seek(startMilliseconds / 1000);
	}

	function lineColor(index: number): string {
		if (!synced || index === activeIndex) {
			return colors.text;
		}
		return inactiveLineColor;
	}

	function lineHover(): { color?: string } {
		if (!synced) {
			return {};
		}
		return { color: hoveredLineColor };
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

{#snippet lrclibToggle()}
	<View style={{ maxWidth: 448 }}>
		<Checkbox
			size="sm"
			label="Find synced lyrics on LRCLIB"
			supportingText="When YouTube has no timed lyrics, the song's title, artist, album and length are sent to lrclib.net."
			checked={lyricsSettings.allowLRCLIB}
			onchange={(allow) => lyricsSettings.setAllowLRCLIB(allow)}
		/>
	</View>
{/snippet}

{#snippet message(text: string)}
	<View style={{ flex: 1, align: 'center', justify: 'center', gap: 24 }}>
		<Label size="xlarge" weight="semibold" tone="dim">{text}</Label>
		{@render lrclibToggle()}
	</View>
{/snippet}

{#if lyricsView.status === 'loading'}
	<View style={{ flex: 1, align: 'center', justify: 'center' }}>
		<LoadingSpinner size={28} color={appearance.accent} />
	</View>
{:else if lyricsView.status === 'ready' && lyricsView.lyrics}
	<Scroll
		bind:this={scroller}
		style={{ flex: 1, minHeight: 0 }}
		onscroll={handleScroll}
		onlayout={(layout) => (viewportHeight = layout.height)}
	>
		<View
			style={{
				gap: lineGap,
				paddingTop: edgePadding,
				paddingBottom: edgePadding,
				paddingRight: 24
			}}
		>
			{#each lines as line, index (index)}
				<View onresize={(layout) => recordLineHeight(index, layout.height)}>
					{#if line.text === ''}
						<View style={{ height: blankLineHeight }} />
					{:else}
						<Button
							style={{
								color: lineColor(index),
								transition: hoverTransition,
								alignSelf: 'stretch'
							}}
							hover={lineHover()}
							disabled={!synced}
							onpress={() => seekToLine(line.startMs)}
						>
							<Label
								size="title"
								weight="bold"
								style={{ flex: 1, minWidth: 0, fontSize: lineFontSize, lineHeight }}
							>
								{line.text}
							</Label>
						</Button>
					{/if}
				</View>
			{/each}
			{#if lyricsView.lyrics.source}
				<View style={{ paddingTop: 32 }}>
					<Label size="small" tone="dim">Source: {lyricsView.lyrics.source}</Label>
				</View>
			{/if}
			{#if offersLRCLIB()}
				<View style={{ paddingTop: 16 }}>{@render lrclibToggle()}</View>
			{/if}
		</View>
	</Scroll>
{:else if lyricsView.status === 'failed'}
	{@render message('Lyrics could not be loaded')}
{:else}
	{@render message('No lyrics for this song')}
{/if}
