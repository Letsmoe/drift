import { searchSongs } from '#lib/api/catalog.ts';
import type { Track } from '#lib/api/music.ts';
import { describeError } from '#lib/format.ts';

export type SearchStatus = 'idle' | 'loading' | 'done' | 'failed';

export class SongSearch {
	tracks = $state<Track[]>([]);
	status = $state<SearchStatus>('idle');
	errorMessage = $state('');

	// Responses can arrive out of order; only the latest request may update state.
	#latestRequestId = 0;

	async run(query: string): Promise<void> {
		const trimmedQuery = query.trim();
		if (trimmedQuery === '') {
			this.status = 'idle';
			return;
		}

		this.#latestRequestId += 1;
		const requestId = this.#latestRequestId;
		this.status = 'loading';

		try {
			const tracks = await searchSongs(trimmedQuery);
			if (requestId === this.#latestRequestId) {
				this.#showResults(tracks);
			}
		} catch (error) {
			if (requestId === this.#latestRequestId) {
				this.#showError(error);
			}
		}
	}

	#showResults(tracks: Track[]): void {
		this.tracks = tracks;
		this.status = 'done';
	}

	#showError(error: unknown): void {
		this.errorMessage = describeError(error);
		this.status = 'failed';
	}
}
