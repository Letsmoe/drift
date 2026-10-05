import { Lyrics } from '#lib/wailsjs/go/main/App.js';
import type { Lyrics as LyricsPage } from '#lib/api/music.ts';

export type { LyricLine, Lyrics } from '#lib/api/music.ts';

/** YouTube Music's lyrics for a track, or null when it has none. */
export async function lyrics(videoId: string): Promise<LyricsPage | null> {
	const found = await Lyrics(videoId);
	if (!found) {
		return null;
	}
	return found;
}
