import {
	QueueAdd,
	QueueAdvance,
	QueueClear,
	QueueJump,
	QueueMove,
	QueueNext,
	QueuePlay,
	QueuePlayNext,
	QueuePrevious,
	QueueRemove,
	QueueSetAutoplay,
	QueueSetMagicShuffle,
	QueueSetRepeat,
	QueueSetShuffle,
	QueueState as fetchQueueState
} from '#lib/wailsjs/go/main/App.js';
import { EventsOn } from '#lib/wailsjs/runtime/runtime.js';
import { queue, type main } from '#lib/wailsjs/go/models.ts';
import type { ItemKind, Track } from '#lib/api/music.ts';

export type QueueState = queue.State;
export type QueueEntry = queue.Entry;
export type QueueStep = main.QueueStep;
export type RepeatMode = 'off' | 'all' | 'one';

/** The state of a queue with nothing in it. */
export function emptyQueueState(): QueueState {
	return queue.State.createFrom({
		entries: [],
		index: -1,
		shuffle: false,
		magicShuffle: false,
		repeat: 'off',
		autoplay: true,
		context: ''
	});
}

const queueEventName = 'queue';

export function queueState(): Promise<QueueState> {
	return fetchQueueState();
}

/** Subscribes to every queue change, including autoplay extensions. Returns the unsubscribe function. */
export function onQueueChange(callback: (state: QueueState) => void): () => void {
	return EventsOn(queueEventName, callback);
}

export function queuePlay(
	tracks: Track[],
	kind: ItemKind,
	startIndex: number,
	playContext: string
): Promise<QueueState> {
	return QueuePlay(tracks, kind, startIndex, playContext);
}

export function queuePlayNext(tracks: Track[]): Promise<QueueState> {
	return QueuePlayNext(tracks);
}

export function queueAdd(tracks: Track[]): Promise<QueueState> {
	return QueueAdd(tracks);
}

export function queueRemove(entryId: number): Promise<QueueState> {
	return QueueRemove(entryId);
}

export function queueMove(entryId: number, newIndex: number): Promise<QueueState> {
	return QueueMove(entryId, newIndex);
}

export function queueJump(entryId: number): Promise<QueueState> {
	return QueueJump(entryId);
}

export function queueClear(): Promise<QueueState> {
	return QueueClear();
}

/** User skip. `moved` is false when the queue has nowhere to go. */
export function queueNext(): Promise<QueueStep> {
	return QueueNext();
}

/** The track ended. `moved` is false at the end of the queue; with repeat one it is true and the index is unchanged. */
export function queueAdvance(): Promise<QueueStep> {
	return QueueAdvance();
}

export function queuePrevious(): Promise<QueueState> {
	return QueuePrevious();
}

export function queueSetShuffle(enabled: boolean): Promise<QueueState> {
	return QueueSetShuffle(enabled);
}

/** Magic shuffle: shuffle plus recommendations mixed in between the queued songs. */
export function queueSetMagicShuffle(enabled: boolean): Promise<QueueState> {
	return QueueSetMagicShuffle(enabled);
}

export function queueSetRepeat(mode: RepeatMode): Promise<QueueState> {
	return QueueSetRepeat(mode);
}

export function queueSetAutoplay(enabled: boolean): Promise<QueueState> {
	return QueueSetAutoplay(enabled);
}
