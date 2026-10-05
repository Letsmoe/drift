import { CoverColors } from '#lib/wailsjs/go/main/App.js';
import type { palette } from '#lib/wailsjs/go/models.ts';

export type Palette = palette.Palette;

export function coverColors(url: string): Promise<Palette> {
	return CoverColors(url);
}
