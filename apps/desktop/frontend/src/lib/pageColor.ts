const saturation = 0.4;
const lightness = 0.3;

function channelHex(value: number): string {
	return Math.round(value * 255)
		.toString(16)
		.padStart(2, '0');
}

// CSS hsl() to rgb, for a hue in degrees.
function hslChannel(hue: number, offset: number): number {
	const position = (offset + hue / 30) % 12;
	const amplitude = saturation * Math.min(lightness, 1 - lightness);
	return lightness - amplitude * Math.max(-1, Math.min(position - 3, 9 - position, 1));
}

/** A stable, muted #rrggbb colour for a page's header gradient, derived from its title. */
export function gradientColorFor(seed: string): string {
	let hash = 0;
	for (const character of seed) {
		hash = (hash * 31 + character.charCodeAt(0)) % 360;
	}
	const red = hslChannel(hash, 0);
	const green = hslChannel(hash, 8);
	const blue = hslChannel(hash, 4);
	return `#${channelHex(red)}${channelHex(green)}${channelHex(blue)}`;
}
