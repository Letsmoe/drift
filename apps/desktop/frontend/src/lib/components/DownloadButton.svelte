<script lang="ts">
	import { Button, Icon } from '@neoworks/svelte-gpui/ui';
	import type { ItemKind, Track } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import ProgressRing from '#lib/components/ProgressRing.svelte';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { downloadFraction } from '#lib/downloads/downloadProgress.ts';
	import {
		arrowCircleDown,
		arrowCircleDownFill,
		clock,
		pauseCircle,
		warningCircleFill
	} from '#lib/icons.ts';
	import { colors, hoverTransition } from '#lib/theme.ts';

	let {
		track,
		kind = 'song',
		size = 18,
		revealOnHover = true,
		rowHovered = false
	}: {
		track: Track;
		kind?: ItemKind;
		size?: number;
		/** In rows, idle and finished downloads only show while the row is hovered. */
		revealOnHover?: boolean;
		/** Whether the pointer is over the row this button sits in. */
		rowHovered?: boolean;
	} = $props();

	const status = $derived(downloads.statusOf(track.videoId));
	const state = $derived(downloads.stateOf(track.videoId));
	// Work in progress and failures always show.
	const alwaysVisible = $derived(!revealOnHover || (state !== '' && state !== 'done'));
	const opacity = $derived(opacityFor(alwaysVisible || rowHovered));
	const color = $derived(colorFor(state));

	function opacityFor(visible: boolean): number {
		if (visible) {
			return 1;
		}
		return 0;
	}

	function colorFor(current: string): string {
		if (current === 'done') {
			return appearance.accent;
		}
		if (current === 'failed') {
			return colors.red;
		}
		return appearance.dim;
	}

	function hoverColor(current: string): string {
		if (current === 'done') {
			return appearance.accent;
		}
		if (current === 'failed') {
			return colors.red;
		}
		return colors.text;
	}

	function iconFor(current: string): string {
		switch (current) {
			case 'queued':
				return clock;
			case 'paused':
				return pauseCircle;
			case 'failed':
				return warningCircleFill;
			case 'done':
				return arrowCircleDownFill;
			default:
				return arrowCircleDown;
		}
	}

	function act(): void {
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

<Button
	style={{ color, opacity, transition: hoverTransition, align: 'center', justify: 'center' }}
	hover={{ color: hoverColor(state) }}
	onpress={act}
>
	{#if state === 'downloading'}
		<ProgressRing fraction={downloadFraction(status)} {size} />
	{:else}
		<Icon svg={iconFor(state)} {size} />
	{/if}
</Button>
