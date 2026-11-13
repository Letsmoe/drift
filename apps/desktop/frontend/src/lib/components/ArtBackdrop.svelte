<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { artWash } from '#lib/layout/artTint.ts';
	import { nowPlayingPalette } from '#lib/palette/coverPalette.svelte.ts';

	// The old backdrop was four gradients of the cover's colours; the host paints flat colours, so
	// this is a wash of the main one, placed first inside a positioned parent. It stays while the
	// colours ease out after the cover is gone.
	const visible = $derived(nowPlayingPalette.eased.tint > 0.001);
</script>

{#if visible}
	<View style={{ position: 'absolute', inset: 0, background: artWash(1, 0.4) }} />
{/if}
