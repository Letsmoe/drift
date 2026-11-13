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

/** The page header's muted colour at a hue in degrees, as #rrggbb. */
function mutedColorAt(hue: number): string {
	const red = hslChannel(hue, 0);
	const green = hslChannel(hue, 8);
	const blue = hslChannel(hue, 4);
	return `#${channelHex(red)}${channelHex(green)}${channelHex(blue)}`;
}

/** A stable, muted #rrggbb colour for a page's header gradient, derived from its title. */
export function gradientColorFor(seed: string): string {
	let hash = 0;
	for (const character of seed) {
		hash = (hash * 31 + character.charCodeAt(0)) % 360;
	}
	return mutedColorAt(hash);
}

/** The hue of a #rrggbb colour in degrees, or undefined for a grey, which has none. */
function hueOf(color: string): number | undefined {
	const red = parseInt(color.slice(1, 3), 16) / 255;
	const green = parseInt(color.slice(3, 5), 16) / 255;
	const blue = parseInt(color.slice(5, 7), 16) / 255;
	const highest = Math.max(red, green, blue);
	const range = highest - Math.min(red, green, blue);
	if (range < 0.04) {
		return undefined;
	}
	if (highest === red) {
		return (60 * ((green - blue) / range) + 360) % 360;
	}
	if (highest === green) {
		return 60 * ((blue - red) / range) + 120;
	}
	return 60 * ((red - green) / range) + 240;
}

/**
 * The header gradient's colour for a page with a cover: the hue of the cover's vibrant colour, as muted as
 * the title-derived one, so the page matches its cover. Undefined when the cover has no hue to take.
 */
export function gradientColorFromCover(vibrant: string): string | undefined {
	if (!vibrant.startsWith('#') || vibrant.length !== 7) {
		return undefined;
	}
	const hue = hueOf(vibrant);
	if (hue === undefined) {
		return undefined;
	}
	return mutedColorAt(hue);
}
