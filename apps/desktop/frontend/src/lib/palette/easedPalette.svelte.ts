import type { Color } from '@neoworks/svelte-gpui/ui';
import { Tween } from 'svelte/motion';
import { cubicInOut } from 'svelte/easing';
import type { Palette } from '#lib/api/palette.ts';
import { mixColors } from '#lib/palette/colorMix.ts';
import { colors as themeColors } from '#lib/theme.ts';

// The old shell eased all of its --art-* colours over this long.
const easeDuration = 1200;
const slotCount = 5;

/** What the cover gives the UI at one moment: how much it tints, and the colours it tints with. */
interface TintState {
	/** 0 without a cover, 1 with one. */
	tint: number;
	vibrant: Color;
	accent: Color;
	/** The vibrant colour followed by the cover's other colours, repeated to fill every slot. */
	slots: Color[];
}

const untintedState: TintState = {
	tint: 0,
	// Without a cover the washes are the glass colour, and the accent is the theme's action colour.
	vibrant: themeColors.surface,
	accent: themeColors.action,
	slots: new Array<Color>(slotCount).fill(themeColors.surface)
};

function isHexColor(value: string): boolean {
	return value.startsWith('#') && value.length === 7;
}

function slotColors(palette: Palette): Color[] {
	const candidates = [palette.vibrant, ...palette.colors].filter(isHexColor);
	const filled: Color[] = [];
	for (let slot = 0; slot < slotCount; slot += 1) {
		filled.push(candidates[slot % candidates.length]);
	}
	return filled;
}

function stateFor(palette: Palette | null): TintState {
	if (palette === null || !isHexColor(palette.vibrant) || !isHexColor(palette.accent)) {
		return untintedState;
	}
	return {
		tint: 1,
		vibrant: palette.vibrant,
		accent: palette.accent,
		slots: slotColors(palette)
	};
}

function mixStates(from: TintState, to: TintState, amount: number): TintState {
	if (amount >= 1) {
		return to;
	}
	const slots: Color[] = [];
	for (let slot = 0; slot < slotCount; slot += 1) {
		slots.push(mixColors(from.slots[slot], to.slots[slot], amount));
	}
	return {
		tint: from.tint + (to.tint - from.tint) * amount,
		vibrant: mixColors(from.vibrant, to.vibrant, amount),
		accent: mixColors(from.accent, to.accent, amount),
		slots
	};
}

/**
 * The cover's colours as the UI shows them: they ease towards each new cover over 1.2 s, so everything
 * that reads them follows. One tween drives the whole palette, so a frame costs one mix per colour here
 * and nothing per element that reads it.
 */
export class EasedPalette {
	#from = $state.raw<TintState>(untintedState);
	#to = $state.raw<TintState>(untintedState);
	#progress = new Tween(1, { duration: easeDuration, easing: cubicInOut });
	#targetPalette: Palette | null = null;
	#current = $derived(mixStates(this.#from, this.#to, this.#progress.current));

	/** 0 without a cover and 1 with one, easing in between. */
	tint = $derived(this.#current.tint);
	vibrant = $derived(this.#current.vibrant);
	accent = $derived(this.#current.accent);

	/** One of the cover's colours; slot 0 is the vibrant one. */
	slot(index: number): Color {
		return this.#current.slots[index % slotCount];
	}

	/** Starts easing from what is shown now to the colours of `palette` (or to none). */
	retarget(palette: Palette | null): void {
		if (palette === this.#targetPalette) {
			return;
		}
		this.#targetPalette = palette;
		this.#from = this.#current;
		this.#to = stateFor(palette);
		this.#progress.set(0, { duration: 0 });
		void this.#progress.set(1);
	}
}
