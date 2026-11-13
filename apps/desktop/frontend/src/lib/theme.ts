// Drift's look as plain data: the colours of its dark theme, the type scale and the style objects that
// several components share. Colours that change with the playing cover (accent, secondary text) are in
// `appearance.svelte.ts`, because they are reactive.

import type { Color, Shadow, Style, Transition } from '@neoworks/svelte-gpui/ui';
import { setThemeMode, themes } from '@neoworks/svelte-gpui-ui';

export const colors = {
	canvas: '#141416',
	elevated: '#1c1c1e',
	surface: '#1c1c1e',
	raised: '#262628',
	hover: '#2e2e31',
	pressed: '#3a3a3e',
	input: '#0e0e10',

	line: '#29292b',
	lineStrong: '#3f3f46',
	lineFaint: 'rgba(255, 255, 255, 0.06)',

	text: '#fafafa',
	muted: '#a1a1aa',
	dim: '#71717a',
	faint: '#52525b',
	inverse: '#0b0b0d',

	action: '#fafafa',
	actionHover: '#e4e4e7',
	onAction: '#0b0b0d',

	green: '#4ade80',
	red: '#f87171',
	amber: '#fbbf24',
	blue: '#60a5fa',
	violet: '#a78bfa',
	pink: '#f472b6',
	redSoft: '#3a2325',

	scrim: 'rgba(0, 0, 0, 0.66)',
	black: '#000000',
	white: '#ffffff'
} as const satisfies Record<string, Color>;

/** Panels let the window's backdrop show through: the surface at 68%. */
export const glassBackground: Color = 'rgba(28, 28, 30, 0.68)';

export const radii = {
	small: 6,
	medium: 8,
	large: 12,
	extraLarge: 16,
	huge: 20,
	full: 9999
} as const;

export const hoverTransition: Transition = { duration: 120, easing: 'ease-out' };

/** Font sizes in pixels, named after what they are used for. */
export const fontSizes = {
	tiny: 12,
	small: 13,
	body: 14,
	medium: 16,
	large: 18,
	xlarge: 20,
	heading: 24,
	title: 30,
	display: 36,
	pageTitle: 60,
	pageTitleWide: 72,
	hero: 96
} as const;

export type TextSize = keyof typeof fontSizes;

/** Line height as a multiple of the font size; the old app inherited 1.5 and tightened its big titles. */
const titleLineHeights: Partial<Record<TextSize, number>> = {
	pageTitle: 1.15,
	pageTitleWide: 1.15,
	hero: 1.05
};

/** The window width from which the old layout used its bigger titles (Tailwind's `xl` breakpoint). */
export const wideWindowWidth = 1280;
const defaultLineHeight = 1.5;

export function textSizeStyle(size: TextSize): Style {
	const fontSize = fontSizes[size];
	let multiple = titleLineHeights[size];
	if (multiple === undefined) {
		multiple = defaultLineHeight;
	}
	return { fontSize, lineHeight: Math.round(fontSize * multiple) };
}

/** `color` with `alpha` (0 to 1) as an rgba string. Only #rrggbb colours are accepted. */
export function withAlpha(color: Color, alpha: number): Color {
	const red = parseInt(color.slice(1, 3), 16);
	const green = parseInt(color.slice(3, 5), 16);
	const blue = parseInt(color.slice(5, 7), 16);
	return `rgba(${red}, ${green}, ${blue}, ${alpha})`;
}

/** A thin ring in the accent colour around the tile of the playing track, standing in for `ring-1 ring-action/40`. */
export function accentRing(accent: Color): Shadow {
	return { x: 0, y: 0, blur: 1, spread: 1, color: withAlpha(accent, 0.4) };
}

/** The floating card of menus and popovers. */
export const floatingPanel: Style = {
	background: colors.elevated,
	borderWidth: 1,
	borderColor: colors.line,
	radius: radii.large,
	padding: 4,
	shadow: { y: 12, blur: 32, color: 'rgba(0, 0, 0, 0.55)' },
	color: colors.text
};

/** A raised tile such as a song tile or a podcast card. */
export const raisedTile: Style = {
	background: colors.raised,
	radius: radii.medium,
	transition: hoverTransition
};

export const raisedTileHover: Style = { background: colors.hover };

/** Horizontal gutter of the page content, as `px-3` and `px-6` were. */
export const gutter = { narrow: 12, wide: 24 } as const;

/**
 * The NeoWorks kit draws from its own theme. Drift's dark colours are put into the kit's dark theme once,
 * so a kit `Button` or `Dialog` matches the components written here.
 */
function installDriftColorsInKit(): void {
	Object.assign(themes.dark.colors, {
		canvas: colors.canvas,
		surface: colors.surface,
		sunken: colors.input,
		hover: colors.hover,
		pressed: colors.pressed,
		line: colors.line,
		lineStrong: colors.lineStrong,
		text: colors.text,
		muted: colors.muted,
		faint: colors.dim,
		fill: colors.action,
		fillHover: colors.actionHover,
		onFill: colors.onAction,
		accent: colors.action,
		accentHover: colors.actionHover,
		onAccent: colors.onAction,
		danger: colors.red,
		dangerSoft: colors.redSoft,
		dangerSoftHover: '#46282a'
	});
	setThemeMode('dark');
}

installDriftColorsInKit();
