// Mixes two colours in Oklab, where the middle of a transition keeps its lightness instead of sinking
// into the muddy mid-tones that blending sRGB channels gives.

import type { Color } from '@neoworks/svelte-gpui/ui';

interface Rgba {
	red: number;
	green: number;
	blue: number;
	alpha: number;
}

interface Oklab {
	lightness: number;
	greenRed: number;
	blueYellow: number;
}

const rgbaPattern = /^rgba?\(\s*([\d.]+)\s*,\s*([\d.]+)\s*,\s*([\d.]+)\s*(?:,\s*([\d.]+)\s*)?\)$/;

function parseColor(color: Color): Rgba {
	if (color.startsWith('#') && color.length === 7) {
		return {
			red: parseInt(color.slice(1, 3), 16),
			green: parseInt(color.slice(3, 5), 16),
			blue: parseInt(color.slice(5, 7), 16),
			alpha: 1
		};
	}
	const match = rgbaPattern.exec(color);
	if (match === null) {
		throw new Error(`mixColors reads #rrggbb and rgb() colours, not "${color}"`);
	}
	let alpha = 1;
	if (match[4] !== undefined) {
		alpha = parseFloat(match[4]);
	}
	return {
		red: parseFloat(match[1]),
		green: parseFloat(match[2]),
		blue: parseFloat(match[3]),
		alpha
	};
}

function toLinear(channel: number): number {
	const unit = channel / 255;
	if (unit <= 0.04045) {
		return unit / 12.92;
	}
	return Math.pow((unit + 0.055) / 1.055, 2.4);
}

function fromLinear(linear: number): number {
	const clamped = Math.min(1, Math.max(0, linear));
	if (clamped <= 0.0031308) {
		return clamped * 12.92 * 255;
	}
	return (1.055 * Math.pow(clamped, 1 / 2.4) - 0.055) * 255;
}

function toOklab(color: Rgba): Oklab {
	const red = toLinear(color.red);
	const green = toLinear(color.green);
	const blue = toLinear(color.blue);
	const long = Math.cbrt(0.4122214708 * red + 0.5363325363 * green + 0.0514459929 * blue);
	const medium = Math.cbrt(0.2119034982 * red + 0.6806995451 * green + 0.1073969566 * blue);
	const short = Math.cbrt(0.0883024619 * red + 0.2817188376 * green + 0.6299787005 * blue);
	return {
		lightness: 0.2104542553 * long + 0.793617785 * medium - 0.0040720468 * short,
		greenRed: 1.9779984951 * long - 2.428592205 * medium + 0.4505937099 * short,
		blueYellow: 0.0259040371 * long + 0.7827717662 * medium - 0.808675766 * short
	};
}

function fromOklab(color: Oklab, alpha: number): Rgba {
	const long = Math.pow(
		color.lightness + 0.3963377774 * color.greenRed + 0.2158037573 * color.blueYellow,
		3
	);
	const medium = Math.pow(
		color.lightness - 0.1055613458 * color.greenRed - 0.0638541728 * color.blueYellow,
		3
	);
	const short = Math.pow(
		color.lightness - 0.0894841775 * color.greenRed - 1.291485548 * color.blueYellow,
		3
	);
	return {
		red: fromLinear(4.0767416621 * long - 3.3077115913 * medium + 0.2309699292 * short),
		green: fromLinear(-1.2684380046 * long + 2.6097574011 * medium - 0.3413193965 * short),
		blue: fromLinear(-0.0041960863 * long - 0.7034186147 * medium + 1.707614701 * short),
		alpha
	};
}

function hexByte(channel: number): string {
	return Math.round(channel).toString(16).padStart(2, '0');
}

function formatColor(color: Rgba): Color {
	if (color.alpha >= 1) {
		return `#${hexByte(color.red)}${hexByte(color.green)}${hexByte(color.blue)}`;
	}
	const alpha = Math.round(color.alpha * 1000) / 1000;
	return `rgba(${Math.round(color.red)}, ${Math.round(color.green)}, ${Math.round(color.blue)}, ${alpha})`;
}

function mixNumbers(from: number, to: number, amount: number): number {
	return from + (to - from) * amount;
}

/** `from` at 0 and `to` at 1, in Oklab, with the opacity mixed linearly. Accepts #rrggbb and rgb()/rgba(). */
export function mixColors(from: Color, to: Color, amount: number): Color {
	if (amount <= 0) {
		return from;
	}
	if (amount >= 1) {
		return to;
	}
	const fromRgba = parseColor(from);
	const toRgba = parseColor(to);
	const fromLab = toOklab(fromRgba);
	const toLab = toOklab(toRgba);
	const mixed: Oklab = {
		lightness: mixNumbers(fromLab.lightness, toLab.lightness, amount),
		greenRed: mixNumbers(fromLab.greenRed, toLab.greenRed, amount),
		blueYellow: mixNumbers(fromLab.blueYellow, toLab.blueYellow, amount)
	};
	return formatColor(fromOklab(mixed, mixNumbers(fromRgba.alpha, toRgba.alpha, amount)));
}
