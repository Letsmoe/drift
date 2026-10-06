import { Lyrics } from '#lib/wailsjs/go/main/App.js';
import type { Lyrics as LyricsPage, Track } from '#lib/api/music.ts';

export type { LyricLine, Lyrics } from '#lib/api/music.ts';

/**
 * The lyrics of a track, or null when no source has any. YouTube comes first;
 * allowLRCLIB lets the core ask lrclib.net when YouTube has no synced lyrics.
 */
export async function lyrics(track: Track, allowLRCLIB: boolean): Promise<LyricsPage | null> {
	const found = await Lyrics(track, allowLRCLIB);
	if (!found) {
		return null;
	}
	return found;
}
