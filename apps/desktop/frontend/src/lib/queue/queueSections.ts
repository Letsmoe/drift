import type { QueueEntry, QueueState } from '#lib/api/queue.ts';

/** An upcoming entry with its absolute position in the queue, which `player.move` needs. */
export type UpcomingEntry = { entry: QueueEntry; index: number };

export type QueueSections = {
	nowPlaying: QueueEntry | undefined;
	queued: UpcomingEntry[];
	autoplay: UpcomingEntry[];
};

export function splitQueue(state: QueueState): QueueSections {
	const queued: UpcomingEntry[] = [];
	const autoplay: UpcomingEntry[] = [];
	for (let index = state.index + 1; index < state.entries.length; index += 1) {
		const entry = state.entries[index];
		if (entry.autoplay) {
			autoplay.push({ entry, index });
		} else {
			queued.push({ entry, index });
		}
	}
	return { nowPlaying: state.entries[state.index], queued, autoplay };
}
