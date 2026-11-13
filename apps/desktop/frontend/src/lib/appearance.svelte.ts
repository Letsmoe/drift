import type { Color } from '@neoworks/svelte-gpui/ui';
import { nowPlayingPalette } from '#lib/palette/coverPalette.svelte.ts';
import { colors } from '#lib/theme.ts';

// With a cover playing, secondary text turns translucent white because the theme's greys disappear on mid-tone tints.
const tintedMuted: Color = 'rgba(255, 255, 255, 0.78)';
const tintedDim: Color = 'rgba(255, 255, 255, 0.62)';
const tintedFaint: Color = 'rgba(255, 255, 255, 0.45)';

/** The colours that follow the cover of the playing track: the accent and the secondary text. */
class Appearance {
	tinted = $derived(nowPlayingPalette.hasColors());
	accent = $derived(this.#accentColor());
	accentHover = $derived(this.#accentHoverColor());
	onAccent = colors.onAction;
	muted = $derived(this.#pick(tintedMuted, colors.muted));
	dim = $derived(this.#pick(tintedDim, colors.dim));
	faint = $derived(this.#pick(tintedFaint, colors.faint));

	#pick(tinted: Color, plain: Color): Color {
		if (this.tinted) {
			return tinted;
		}
		return plain;
	}

	#accentColor(): Color {
		if (!this.tinted) {
			return colors.action;
		}
		return nowPlayingPalette.colors!.accent;
	}

	#accentHoverColor(): Color {
		if (!this.tinted) {
			return colors.actionHover;
		}
		return nowPlayingPalette.colors!.accent;
	}
}

export const appearance = new Appearance();
