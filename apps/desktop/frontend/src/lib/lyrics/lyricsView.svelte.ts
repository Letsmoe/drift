import { lyrics as fetchLyrics, type Lyrics } from '#lib/api/lyrics.ts';
import type { Track } from '#lib/api/music.ts';
import { lyricsSettings } from '#lib/lyrics/lyricsSettings.svelte.ts';

export type LyricsStatus = 'idle' | 'loading' | 'ready' | 'missing' | 'failed';

/** Lyrics of the playing track, cached per video ID and LRCLIB setting. */
class LyricsView {
	status = $state<LyricsStatus>('idle');
	lyrics = $state<Lyrics | null>(null);

	private cache = new Map<string, Lyrics | null>();
	private requestedKey = '';

	/** Loads the lyrics for track; a later call for another track wins over a slower earlier one. */
	async load(track: Track): Promise<void> {
		const allowLRCLIB = lyricsSettings.allowLRCLIB;
		const key = `${track.videoId}:${allowLRCLIB}`;
		this.requestedKey = key;
		const cached = this.cache.get(key);
		if (cached !== undefined) {
			this.show(cached);
			return;
		}

		this.lyrics = null;
		this.status = 'loading';
		try {
			const found = await fetchLyrics(track, allowLRCLIB);
			this.cache.set(key, found);
			this.showIfCurrent(key, found);
		} catch {
			this.showFailureIfCurrent(key);
		}
	}

	/** Index of the last line that started at positionSeconds, or -1 for plain lyrics. */
	activeLineIndex(positionSeconds: number): number {
		if (!this.lyrics || !this.lyrics.synced) {
			return -1;
		}
		const lines = this.lyrics.lines;
		const positionMs = positionSeconds * 1000;
		let low = 0;
		let high = lines.length - 1;
		let activeIndex = -1;
		while (low <= high) {
			const middle = Math.floor((low + high) / 2);
			if (lines[middle].startMs <= positionMs) {
				activeIndex = middle;
				low = middle + 1;
			} else {
				high = middle - 1;
			}
		}
		return activeIndex;
	}

	private showIfCurrent(key: string, found: Lyrics | null): void {
		if (key !== this.requestedKey) {
			return;
		}
		this.show(found);
	}

	private showFailureIfCurrent(key: string): void {
		if (key !== this.requestedKey) {
			return;
		}
		this.lyrics = null;
		this.status = 'failed';
	}

	private show(found: Lyrics | null): void {
		this.lyrics = found;
		if (found === null) {
			this.status = 'missing';
			return;
		}
		this.status = 'ready';
	}
}

export const lyricsView = new LyricsView();
