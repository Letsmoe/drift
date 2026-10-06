import { SvelteSet } from 'svelte/reactivity';
import {
	recommendedFeed,
	recommendedRows,
	type RecommendedItem,
	type RecommendedRow
} from '#lib/api/recommend.ts';
import { describeError } from '#lib/format.ts';
import { logClick, logImpression } from '#lib/recommend/impressions.ts';
import { recommendationFeedback } from '#lib/recommend/recommendationFeedback.svelte.ts';

export type SectionStatus = 'loading' | 'done' | 'failed';

const rowLimit = 6;
const itemsPerRow = 12;
const feedSize = 20;
const surface = 'home';

function withoutHiddenItems(row: RecommendedRow): RecommendedRow {
	const items = row.items.filter((item) => !recommendationFeedback.hides(item.track));
	return { ...row, items } as RecommendedRow;
}

/** Personalised home content: the "Made for you" feed and the reason rows, each loading on its own. */
export class Recommendations {
	#loadedFeed = $state<RecommendedItem[]>([]);
	#loadedRows = $state<RecommendedRow[]>([]);
	// Disliked and "Don't recommend" tracks leave the lists at once, not at the next load.
	feed = $derived(this.#loadedFeed.filter((item) => !recommendationFeedback.hides(item.track)));
	rows = $derived(this.#loadedRows.map(withoutHiddenItems).filter((row) => row.items.length > 0));
	feedStatus = $state<SectionStatus>('loading');
	feedError = $state('');
	rowsStatus = $state<SectionStatus>('loading');
	rowsError = $state('');

	#impressedVideoIds = new SvelteSet<string>();

	async load(): Promise<void> {
		await Promise.all([this.#loadFeed(false), this.#loadRows()]);
	}

	/** Re-ranks the feed and reloads the rows. */
	async refresh(): Promise<void> {
		this.#impressedVideoIds.clear();
		await Promise.all([this.#loadFeed(true), this.#loadRows()]);
	}

	/** Logs that an item was shown, once per video. */
	impress(item: RecommendedItem, position: number): void {
		const videoId = item.track.videoId;
		if (this.#impressedVideoIds.has(videoId)) {
			return;
		}
		this.#impressedVideoIds.add(videoId);
		logImpression(videoId, item.source, surface, position);
	}

	click(item: RecommendedItem, position: number): void {
		logClick(item.track.videoId, item.source, surface, position);
	}

	async #loadFeed(refresh: boolean): Promise<void> {
		this.feedStatus = 'loading';
		try {
			this.#loadedFeed = await recommendedFeed(feedSize, 0, refresh);
			this.feedStatus = 'done';
		} catch (error) {
			this.feedError = describeError(error);
			this.feedStatus = 'failed';
		}
	}

	async #loadRows(): Promise<void> {
		this.rowsStatus = 'loading';
		try {
			this.#loadedRows = await recommendedRows(rowLimit, itemsPerRow);
			this.rowsStatus = 'done';
		} catch (error) {
			this.rowsError = describeError(error);
			this.rowsStatus = 'failed';
		}
	}
}
