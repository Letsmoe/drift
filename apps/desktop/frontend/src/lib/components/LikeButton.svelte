<script lang="ts">
	import type { ItemKind, Track } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import IconAction from '#lib/components/IconAction.svelte';
	import { heart, heartFill } from '#lib/icons.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { Pop } from '#lib/motion/pop.svelte.ts';
	import { colors } from '#lib/theme.ts';

	let {
		track,
		kind = 'song',
		size = 20
	}: { track: Track; kind?: ItemKind; size?: number } = $props();

	const liked = $derived(library.isLiked(track.videoId));
	const icon = $derived(iconFor(liked));
	const color = $derived(colorFor(liked));
	const pop = new Pop();

	// The heart pops when it fills; taking a like back needs no fanfare.
	function toggle(): void {
		if (!liked) {
			pop.trigger();
		}
		library.toggleLike(track, kind);
	}

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
	style={{ scale: pop.scale() }}
	onpress={toggle}
/>
