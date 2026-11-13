import type { Thumbnail } from '#lib/api/music.ts';
import { coverColors, type Palette } from '#lib/api/palette.ts';
import { pickThumbnail } from '#lib/format.ts';
import { EasedPalette } from '#lib/palette/easedPalette.svelte.ts';

// The colors come from a small thumbnail; a bigger one finds the same colors slower.
const sampledThumbnailSize = 60;

/** Colors of one cover at a time; keeps the previous colors until the next ones arrive. */
export class CoverPalette {
	colors = $state<Palette | null>(null);
	/** The same colors, easing towards each new cover; what the UI should draw with. */
	eased = new EasedPalette();
	#requestedUrl = '';

	async load(thumbnails: Thumbnail[]): Promise<void> {
		const url = pickThumbnail(thumbnails, sampledThumbnailSize);
		if (url === undefined) {
			this.#requestedUrl = '';
			this.#setColors(null);
			return;
		}
		if (url === this.#requestedUrl) {
			return;
		}
		this.#requestedUrl = url;
		try {
			const result = await coverColors(url);
			if (this.#requestedUrl === url) {
				this.#setColors(result);
			}
		} catch {
			// A cover that can't be read leaves the UI untinted.
			if (this.#requestedUrl === url) {
				this.#setColors(null);
			}
		}
	}

	hasColors(): boolean {
		return this.colors !== null && this.colors.vibrant !== '';
	}

	#setColors(palette: Palette | null): void {
		this.colors = palette;
		this.eased.retarget(palette);
	}
}

/** Colors of the playing track's cover; drives the tint of the shell. */
export const nowPlayingPalette = new CoverPalette();
