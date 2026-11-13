import type { Color } from '@neoworks/svelte-gpui/ui';
import { nowPlayingPalette } from '#lib/palette/coverPalette.svelte.ts';
import { mixColors } from '#lib/palette/colorMix.ts';
import { themes } from '@neoworks/svelte-gpui-ui';
import { colors } from '#lib/theme.ts';

// With a cover playing, secondary text turns translucent white because the theme's greys disappear on mid-tone tints.
const tintedMuted: Color = 'rgba(255, 255, 255, 0.78)';
const tintedDim: Color = 'rgba(255, 255, 255, 0.62)';
const tintedFaint: Color = 'rgba(255, 255, 255, 0.45)';

/**
 * The colours that follow the cover of the playing track: the accent and the secondary text. They read
 * the eased palette, so they change together over the same 1.2 s when the cover does.
 */
class Appearance {
	#tint = $derived(nowPlayingPalette.eased.tint);

	accent = $derived(nowPlayingPalette.eased.accent);
	accentHover = $derived(mixColors(colors.actionHover, nowPlayingPalette.eased.accent, this.#tint));
	onAccent = colors.onAction;
	muted = $derived(mixColors(colors.muted, tintedMuted, this.#tint));
	dim = $derived(mixColors(colors.dim, tintedDim, this.#tint));
	faint = $derived(mixColors(colors.faint, tintedFaint, this.#tint));
}

export const appearance = new Appearance();

/**
 * Kit components fill their buttons, checkboxes and sliders with the theme's `fill` and `accent`. Those two
 * roles read the cover's accent through getters, so a kit component read inside a `$derived` or a style follows
 * the cover like the components written here, easing included.
 */
function followCoverAccentInKit(): void {
	const kitColors = themes.dark.colors;
	const accentRoles = {
		fill: () => appearance.accent,
		accent: () => appearance.accent,
		fillHover: () => appearance.accentHover,
		accentHover: () => appearance.accentHover
	};
	for (const [role, read] of Object.entries(accentRoles)) {
		Object.defineProperty(kitColors, role, { get: read, enumerable: true, configurable: true });
	}
}

followCoverAccentInKit();
