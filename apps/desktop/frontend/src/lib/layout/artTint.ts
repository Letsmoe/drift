import type { Color } from '@neoworks/svelte-gpui/ui';
import { nowPlayingPalette } from '#lib/palette/coverPalette.svelte.ts';
import { glassBackground, withAlpha } from '#lib/theme.ts';

/**
 * A wash of the playing cover's colours, standing in for the old gradients (the host draws flat
 * colours only). `slot` 0 is the vibrant colour, 1 to 4 the palette in order; without a cover
 * playing, the plain glass colour is returned.
 */
export function artTint(slot: number, alpha: number): Color {
	const palette = nowPlayingPalette.colors;
	if (!palette || palette.vibrant === '') {
		return glassBackground;
	}
	const candidates = [palette.vibrant, ...palette.colors];
	const chosen = candidates[slot % candidates.length];
	if (!chosen.startsWith('#') || chosen.length !== 7) {
		return glassBackground;
	}
	return withAlpha(chosen, alpha);
}
