import { tick } from 'svelte';
import type { ItemKind, Track } from '#lib/api/music.ts';
import { resumePosition, startPlay, updatePlay } from '#lib/api/library.ts';
import {
	emptyQueueState,
	onQueueChange,
	queueAdd,
	queueAdvance,
	queueClear,
	queueJump,
	queueMove,
	queueNext,
	queuePlay,
	queuePlayNext,
	queuePrevious,
	queueRemove,
	queueSetAutoplay,
	queueSetMagicShuffle,
	queueSetRepeat,
	queueSetShuffle,
	queueState,
	type QueueEntry,
	type QueueState,
	type RepeatMode
} from '#lib/api/queue.ts';
import { streamUrl } from '#lib/api/music.ts';
import type { RemotePlayback } from '#lib/api/together.ts';
import { describeError } from '#lib/format.ts';

/** Seconds into a track after which Previous restarts it instead of going back. */
const restartThresholdSeconds = 3;
const progressReportIntervalMs = 15_000;
/** Unplayable tracks skipped in a row before giving up, so a dead connection doesn't run through the queue. */
const maximumFailedSkips = 5;
const unplayableMessage = 'This song could not be played.';

/** Where someone else's player was when it arrived, for an entry that is still loading here. */
interface PendingRemoteStart {
	playback: RemotePlayback;
	receivedAtMs: number;
}

/**
 * Playback state shared by the track lists, the queue UI and the player bar. The queue lives
 * in the Go core; this class mirrors it and drives the <audio> element in PlayerBar, which
 * binds to `source`, `paused`, `currentTime`, `duration`, `volume` and `muted`.
 */
export class Player {
	queue = $state<QueueState>(emptyQueueState());
	entry = $derived(this.queue.entries[this.queue.index]);
	track = $derived<Track | null>(this.#trackOf(this.entry));
	source = $state<string | undefined>(undefined);
	paused = $state(true);
	currentTime = $state(0);
	duration = $state(0);
	volume = $state(1);
	muted = $state(false);
	buffering = $state(false);
	errorMessage = $state('');

	// A slow URL lookup must not override a track the user picked afterwards.
	#latestRequestId = 0;
	#loadedEntryId: number | undefined = undefined;
	// The entry `source` belongs to; it lags #loadedEntryId while a URL loads.
	#sourceEntryId: number | undefined = undefined;
	#playIdPromise: Promise<number | null> | null = null;
	#userActionListeners: Array<() => void> = [];
	#pendingRemoteStart: PendingRemoteStart | null = null;
	#retriedEntryId: number | undefined = undefined;
	#failedSkips = 0;

	constructor() {
		onQueueChange((state) => this.#applyState(state));
		queueState().then(
			(state) => this.#adoptInitialState(state),
			() => {}
		);
		setInterval(() => this.#reportWhilePlaying(), progressReportIntervalMs);
	}

	/** Replaces the queue with tracks and starts playing at startIndex. */
	async playTracks(
		tracks: Track[],
		startIndex: number,
		context: string,
		kind: ItemKind = 'song'
	): Promise<void> {
		await this.#run(queuePlay(tracks, kind, startIndex, context));
	}

	async playNext(tracks: Track[]): Promise<void> {
		await this.#run(queuePlayNext(tracks));
	}

	async addToQueue(tracks: Track[]): Promise<void> {
		await this.#run(queueAdd(tracks));
	}

	/** Skips to the next entry; stops playback at the end of the queue. */
	async next(): Promise<void> {
		try {
			const step = await queueNext();
			this.#applyState(step.state);
			if (!step.moved) {
				this.#stopAtEnd();
			}
		} catch (error) {
			this.fail(describeError(error));
		}
	}

	/** Restarts the track when it is past 3 s, otherwise goes back one entry. */
	async previous(): Promise<void> {
		if (this.currentTime > restartThresholdSeconds) {
			this.seek(0);
			return;
		}
		const entryIdBefore = this.entry?.id;
		await this.#run(queuePrevious());
		if (this.entry?.id === entryIdBefore) {
			this.seek(0);
		}
	}

	async jump(entryId: number): Promise<void> {
		await this.#run(queueJump(entryId));
	}

	async remove(entryId: number): Promise<void> {
		await this.#run(queueRemove(entryId));
	}

	async move(entryId: number, newIndex: number): Promise<void> {
		await this.#run(queueMove(entryId, newIndex));
	}

	async clearQueue(): Promise<void> {
		await this.#run(queueClear());
	}

	async toggleShuffle(): Promise<void> {
		await this.#run(queueSetShuffle(!this.queue.shuffle));
	}

	/** Cycles shuffle off, shuffle, magic shuffle (with recommendations mixed in), off. */
	async cycleShuffle(): Promise<void> {
		if (!this.queue.shuffle) {
			await this.#run(queueSetShuffle(true));
			return;
		}
		if (!this.queue.magicShuffle) {
			await this.#run(queueSetMagicShuffle(true));
			return;
		}
		await this.#run(queueSetShuffle(false));
	}

	/** Cycles repeat off, all, one, off. */
	async cycleRepeat(): Promise<void> {
		await this.#run(queueSetRepeat(this.#nextRepeatMode()));
	}

	async toggleAutoplay(): Promise<void> {
		await this.#run(queueSetAutoplay(!this.queue.autoplay));
	}

	/** Called by the <audio> element when the track finished. */
	async trackEnded(): Promise<void> {
		this.#reportPlay(true);
		// The queue event can land before the call returns, so remember the entry from before.
		const endedEntryId = this.entry?.id;
		try {
			const step = await queueAdvance();
			this.#applyState(step.state);
			if (!step.moved) {
				this.#stopAtEnd();
				return;
			}
			if (this.entry?.id === endedEntryId) {
				this.#restartCurrent();
			}
		} catch (error) {
			this.fail(describeError(error));
		}
	}

	/** Plays track from list (the queue becomes the list), or pauses and resumes it when it is already current. */
	playOrToggle(
		track: Track,
		tracks: Track[] = [track],
		startIndex = 0,
		context = 'track',
		kind: ItemKind = 'song'
	): void {
		if (this.isCurrent(track)) {
			this.toggle();
			return;
		}
		this.playTracks(tracks, startIndex, context, kind);
	}

	toggle(): void {
		if (!this.track || this.errorMessage) {
			return;
		}
		if (!this.source) {
			this.#loadEntry(this.entry);
			return;
		}
		if (!this.paused) {
			this.#reportPlay(false);
		}
		this.paused = !this.paused;
		this.#notifyUserAction();
	}

	seek(seconds: number): void {
		this.currentTime = seconds;
		this.#notifyUserAction();
	}

	/** Listens to plays, pauses and seeks by the user. Returns the unsubscribe function. */
	onUserAction(listener: () => void): () => void {
		this.#userActionListeners.push(listener);
		return () => {
			this.#userActionListeners = this.#userActionListeners.filter((other) => other !== listener);
		};
	}

	/** Follows someone else's player (Listen Together) without announcing it as a user action. */
	followRemote(playback: RemotePlayback): void {
		if (this.entry?.id !== playback.entryId || this.errorMessage) {
			return;
		}
		if (this.#sourceEntryId !== playback.entryId) {
			this.#pendingRemoteStart = { playback, receivedAtMs: performance.now() };
			return;
		}
		if (this.paused !== playback.paused) {
			this.paused = playback.paused;
		}
		if (Math.abs(this.currentTime - playback.positionSeconds) > playback.toleranceSeconds) {
			this.currentTime = playback.positionSeconds;
		}
	}

	setVolume(volume: number): void {
		this.volume = volume;
		this.muted = volume === 0;
	}

	toggleMute(): void {
		this.muted = !this.muted;
	}

	isCurrent(track: Track): boolean {
		return this.track?.videoId === track.videoId;
	}

	isPlaying(track: Track): boolean {
		return this.isCurrent(track) && !this.paused;
	}

	/** Called by the <audio> element when the track cannot be played: tries once more, then skips it. */
	async playbackFailed(): Promise<void> {
		const entry = this.entry;
		if (!entry || !this.source) {
			return;
		}
		if (this.#retriedEntryId !== entry.id) {
			this.#retriedEntryId = entry.id;
			await this.#reloadSource();
			return;
		}
		if (this.#failedSkips >= maximumFailedSkips) {
			this.fail(unplayableMessage);
			return;
		}
		this.#failedSkips += 1;
		this.fail(unplayableMessage);
		await this.next();
	}

	/** Called by the <audio> element once playback moves on with data loaded. */
	playbackAdvanced(): void {
		this.buffering = false;
		this.#failedSkips = 0;
	}

	fail(message: string): void {
		this.#reportPlay(false);
		this.errorMessage = message;
		this.buffering = false;
		this.paused = true;
	}

	#trackOf(entry: QueueEntry | undefined): Track | null {
		if (!entry) {
			return null;
		}
		return entry.track;
	}

	#nextRepeatMode(): RepeatMode {
		if (this.queue.repeat === 'off') {
			return 'all';
		}
		if (this.queue.repeat === 'all') {
			return 'one';
		}
		return 'off';
	}

	async #run(request: Promise<QueueState>): Promise<void> {
		try {
			this.#applyState(await request);
		} catch (error) {
			this.fail(describeError(error));
		}
	}

	// A page reload keeps the core's queue; show it without starting playback.
	#adoptInitialState(state: QueueState): void {
		if (this.#loadedEntryId === undefined && this.queue.entries.length === 0) {
			this.queue = state;
		}
	}

	#applyState(state: QueueState): void {
		this.queue = state;
		const entry = this.entry;
		if (entry?.id === this.#loadedEntryId) {
			return;
		}
		if (!entry) {
			this.#unload();
			return;
		}
		this.#loadEntry(entry);
	}

	async #loadEntry(entry: QueueEntry): Promise<void> {
		this.#reportPlay(false);
		this.#latestRequestId += 1;
		const requestId = this.#latestRequestId;
		this.#loadedEntryId = entry.id;

		this.errorMessage = '';
		this.currentTime = 0;
		this.duration = entry.track.durationSeconds;
		this.buffering = true;
		this.#playIdPromise = this.#startPlayReport(entry);

		try {
			const [source, resumeSeconds] = await Promise.all([
				streamUrl(entry.track.videoId),
				this.#lookUpResumeSeconds(entry)
			]);
			if (requestId === this.#latestRequestId) {
				this.#startPlayback(source, resumeSeconds);
			}
		} catch (error) {
			if (requestId === this.#latestRequestId) {
				this.fail(describeError(error));
			}
		}
	}

	// Unsetting the source first makes the <audio> element request the stream again.
	async #reloadSource(): Promise<void> {
		const source = this.source;
		const position = this.currentTime;
		this.source = undefined;
		await tick();
		this.source = source;
		this.currentTime = position;
		this.buffering = true;
		this.paused = false;
	}

	#startPlayback(source: string, resumeSeconds: number): void {
		this.source = source;
		this.#sourceEntryId = this.#loadedEntryId;
		const remoteStart = this.#takePendingRemoteStart();
		if (remoteStart) {
			this.currentTime = remoteStart.positionSeconds;
			this.paused = remoteStart.paused;
			return;
		}
		this.currentTime = resumeSeconds;
		this.paused = false;
	}

	// A remote position that arrived while loading, moved on by the time the load took.
	#takePendingRemoteStart(): RemotePlayback | null {
		const pending = this.#pendingRemoteStart;
		this.#pendingRemoteStart = null;
		if (!pending || pending.playback.entryId !== this.#loadedEntryId) {
			return null;
		}
		if (pending.playback.paused) {
			return pending.playback;
		}
		const elapsedSeconds = (performance.now() - pending.receivedAtMs) / 1000;
		return {
			...pending.playback,
			positionSeconds: pending.playback.positionSeconds + elapsedSeconds
		};
	}

	#notifyUserAction(): void {
		for (const listener of this.#userActionListeners) {
			listener();
		}
	}

	async #lookUpResumeSeconds(entry: QueueEntry): Promise<number> {
		if (entry.kind !== 'episode') {
			return 0;
		}
		try {
			const point = await resumePosition(entry.track.videoId);
			if (point.found) {
				return point.seconds;
			}
		} catch {
			// Resuming is a convenience; start from the beginning instead.
		}
		return 0;
	}

	#unload(): void {
		this.#reportPlay(false);
		this.#latestRequestId += 1;
		this.#loadedEntryId = undefined;
		this.#sourceEntryId = undefined;
		this.source = undefined;
		this.paused = true;
		this.currentTime = 0;
		this.duration = 0;
		this.buffering = false;
		this.errorMessage = '';
	}

	#stopAtEnd(): void {
		this.paused = true;
		this.currentTime = 0;
	}

	// Repeat one: the same entry plays again, as a new play.
	#restartCurrent(): void {
		this.#playIdPromise = this.#startPlayReport(this.entry);
		this.currentTime = 0;
		this.paused = false;
	}

	#startPlayReport(entry: QueueEntry): Promise<number | null> {
		return startPlay(entry.track, entry.kind as ItemKind, this.queue.context).catch((error) => {
			console.warn('startPlay failed', error);
			return null;
		});
	}

	#reportWhilePlaying(): void {
		if (!this.paused) {
			this.#reportPlay(false);
		}
	}

	/** Reports progress of the current play; `completed` also closes it. */
	#reportPlay(completed: boolean): void {
		const playIdPromise = this.#playIdPromise;
		if (!playIdPromise) {
			return;
		}
		if (completed) {
			this.#playIdPromise = null;
		}
		const progressSeconds = Math.floor(this.currentTime);
		playIdPromise.then((playId) => this.#sendProgress(playId, progressSeconds, completed));
	}

	#sendProgress(playId: number | null, progressSeconds: number, completed: boolean): void {
		if (playId === null) {
			return;
		}
		updatePlay(playId, progressSeconds, completed).catch((error) =>
			console.warn('updatePlay failed', error)
		);
	}
}

export const player = new Player();
