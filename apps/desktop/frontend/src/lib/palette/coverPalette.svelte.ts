import type { Thumbnail } from '#lib/api/music.ts';
import { coverColors, type Palette } from '#lib/api/palette.ts';
import { pickThumbnail } from '#lib/format.ts';

// The colors come from a small thumbnail; a bigger one finds the same colors slower.
const sampledThumbnailSize = 60;

/** Colors of one cover at a time; keeps the previous colors until the next ones arrive. */
export class CoverPalette {
	colors = $state<Palette | null>(null);
	#requestedUrl = '';

	async load(thumbnails: Thumbnail[]): Promise<void> {
		const url = pickThumbnail(thumbnails, sampledThumbnailSize);
		if (url === undefined) {
			this.#requestedUrl = '';
			this.colors = null;
			return;
		}
		if (url === this.#requestedUrl) {
			return;
		}
		this.#requestedUrl = url;
		try {
			const result = await coverColors(url);
			if (this.#requestedUrl === url) {
				this.colors = result;
			}
		} catch {
			// A cover that can't be read leaves the UI untinted.
			if (this.#requestedUrl === url) {
				this.colors = null;
			}
		}
	}

	/** CSS custom properties for the colors; empty without colors, so the CSS defaults apply. */
	style(): string {
		if (this.colors === null || this.colors.vibrant === '') {
			return '';
		}
		return `--art-vibrant: ${this.colors.vibrant}; --art-dominant: ${this.colors.dominant};`;
	}
}

/** Colors of the playing track's cover; drives the tint of the shell. */
export const nowPlayingPalette = new CoverPalette();
