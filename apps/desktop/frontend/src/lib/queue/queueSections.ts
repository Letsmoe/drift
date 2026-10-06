import type { QueueEntry, QueueState } from '#lib/api/queue.ts';

/** A queue entry with its absolute position in the queue, which `player.move` needs. */
export type UpcomingEntry = { entry: QueueEntry; index: number };

export type QueueSections = {
	/** Entries played before the current one, oldest first. Unplayed earlier entries are left out. */
	played: UpcomingEntry[];
	nowPlaying: QueueEntry | undefined;
	/** The user's entries, with magic shuffle's recommendations mixed in. */
	queued: UpcomingEntry[];
	autoplay: UpcomingEntry[];
};

export function splitQueue(state: QueueState): QueueSections {
	const played: UpcomingEntry[] = [];
	const queued: UpcomingEntry[] = [];
	const autoplay: UpcomingEntry[] = [];
	for (let index = 0; index < state.index; index += 1) {
		const entry = state.entries[index];
		if (entry.played) {
			played.push({ entry, index });
		}
	}
	for (let index = state.index + 1; index < state.entries.length; index += 1) {
		const entry = state.entries[index];
		if (entry.autoplay) {
			autoplay.push({ entry, index });
		} else {
			queued.push({ entry, index });
		}
	}
	return { played, nowPlaying: state.entries[state.index], queued, autoplay };
}
