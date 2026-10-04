import { recordImpressions, type Impression } from '#lib/api/library.ts';

const flushIntervalMs = 4000;
const maxBatchSize = 50;

/** Collects recommendation impressions and clicks and writes them to the library in batches. */
class ImpressionLogger {
	#pending: Impression[] = [];
	#flushTimer: ReturnType<typeof setTimeout> | undefined = undefined;

	/** An item was shown to the user. */
	logImpression(videoId: string, source: string, surface: string, position: number): void {
		this.#enqueue(videoId, source, surface, position, false);
	}

	/** The user clicked an item that was shown. */
	logClick(videoId: string, source: string, surface: string, position: number): void {
		this.#enqueue(videoId, source, surface, position, true);
	}

	flush(): void {
		clearTimeout(this.#flushTimer);
		this.#flushTimer = undefined;
		if (this.#pending.length === 0) {
			return;
		}
		const batch = this.#pending;
		this.#pending = [];
		recordImpressions(batch).catch((error) => console.warn('recordImpressions failed', error));
	}

	#enqueue(
		videoId: string,
		source: string,
		surface: string,
		position: number,
		clicked: boolean
	): void {
		this.#pending.push({ videoId, source, surface, position, clicked });
		if (this.#pending.length >= maxBatchSize) {
			this.flush();
			return;
		}
		if (this.#flushTimer === undefined) {
			this.#flushTimer = setTimeout(() => this.flush(), flushIntervalMs);
		}
	}
}

export const impressions = new ImpressionLogger();

export function logImpression(
	videoId: string,
	source: string,
	surface: string,
	position: number
): void {
	impressions.logImpression(videoId, source, surface, position);
}

export function logClick(videoId: string, source: string, surface: string, position: number): void {
	impressions.logClick(videoId, source, surface, position);
}
