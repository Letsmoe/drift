import type { Color } from '@neoworks/svelte-gpui/ui';
import { nowPlayingPalette } from '#lib/palette/coverPalette.svelte.ts';
import { glassAlpha, withAlpha } from '#lib/theme.ts';

/**
 * A wash of the playing cover's colours, standing in for the old gradients (the host draws flat
 * colours only). `slot` 0 is the vibrant colour, 1 to 4 the palette in order. Without a cover playing,
 * the plain glass colour is returned; while the cover changes, the colour and the opacity ease.
 */
export function artTint(slot: number, alpha: number): Color {
	const palette = nowPlayingPalette.eased;
	const eased = glassAlpha + (alpha - glassAlpha) * palette.tint;
	return withAlpha(palette.slot(slot), eased);
}

/** Like `artTint`, but laid over a plain background: it fades in from nothing instead of from glass. */
export function artWash(slot: number, alpha: number): Color {
	const palette = nowPlayingPalette.eased;
	return withAlpha(palette.slot(slot), alpha * palette.tint);
}
