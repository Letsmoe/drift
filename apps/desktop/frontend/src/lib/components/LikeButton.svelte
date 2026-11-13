<script lang="ts">
	import type { ItemKind, Track } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import IconAction from '#lib/components/IconAction.svelte';
	import { heart, heartFill } from '#lib/icons.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { colors } from '#lib/theme.ts';

	let {
		track,
		kind = 'song',
		size = 20
	}: { track: Track; kind?: ItemKind; size?: number } = $props();

	const liked = $derived(library.isLiked(track.videoId));
	const icon = $derived(iconFor(liked));
	const color = $derived(colorFor(liked));

	function iconFor(isLiked: boolean): string {
		if (isLiked) {
			return heartFill;
		}
		return heart;
	}

	function colorFor(isLiked: boolean): string {
		if (isLiked) {
			return appearance.accent;
		}
		return appearance.muted;
	}

	function hoverColorFor(isLiked: boolean): string {
		if (isLiked) {
			return appearance.accent;
		}
		return colors.text;
	}
</script>

<IconAction
	{icon}
	iconSize={size}
	boxSize={32}
	{color}
	hoverColor={hoverColorFor(liked)}
	hoverBackground={colors.hover}
	onpress={() => library.toggleLike(track, kind)}
/>
