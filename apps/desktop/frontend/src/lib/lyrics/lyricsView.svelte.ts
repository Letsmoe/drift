import { lyrics as fetchLyrics, type Lyrics } from '#lib/api/lyrics.ts';

export type LyricsStatus = 'idle' | 'loading' | 'ready' | 'missing' | 'failed';

/** Lyrics of the track the full-screen player shows, cached per video ID. */
class LyricsView {
	status = $state<LyricsStatus>('idle');
	lyrics = $state<Lyrics | null>(null);

	private cache = new Map<string, Lyrics | null>();
	private requestedVideoId = '';

	/** Loads the lyrics for videoId; a later call for another track wins over a slower earlier one. */
	async load(videoId: string): Promise<void> {
		this.requestedVideoId = videoId;
		const cached = this.cache.get(videoId);
		if (cached !== undefined) {
			this.show(cached);
			return;
		}

		this.lyrics = null;
		this.status = 'loading';
		try {
			const found = await fetchLyrics(videoId);
			this.cache.set(videoId, found);
			this.showIfCurrent(videoId, found);
		} catch {
			this.showFailureIfCurrent(videoId);
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

	private showIfCurrent(videoId: string, found: Lyrics | null): void {
		if (videoId !== this.requestedVideoId) {
			return;
		}
		this.show(found);
	}

	private showFailureIfCurrent(videoId: string): void {
		if (videoId !== this.requestedVideoId) {
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
