import { CoverColors } from '#lib/backend/generated/App.ts';
import type { palette } from '#lib/backend/generated/models.ts';

export type Palette = palette.Palette;

export function coverColors(url: string): Promise<Palette> {
	return CoverColors(url);
}
