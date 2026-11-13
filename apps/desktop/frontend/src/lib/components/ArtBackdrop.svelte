<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { artTint } from '#lib/layout/artTint.ts';
	import { nowPlayingPalette } from '#lib/palette/coverPalette.svelte.ts';
	import { hoverTransition } from '#lib/theme.ts';

	// The old backdrop was four gradients of the cover's colours; the host paints flat colours, so
	// this is a wash of the main one, placed first inside a positioned parent.
	const tinted = $derived(nowPlayingPalette.hasColors());
</script>

{#if tinted}
	<View
		style={{
			position: 'absolute',
			inset: 0,
			background: artTint(1, 0.4),
			transition: { ...hoverTransition, duration: 600 }
		}}
	/>
{/if}
