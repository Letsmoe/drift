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
import { describeError } from '#lib/format.ts';

/** Seconds into a track after which Previous restarts it instead of going back. */
const restartThresholdSeconds = 3;
const progressReportIntervalMs = 15_000;

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
	#playIdPromise: Promise<number | null> | null = null;

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
	}

	seek(seconds: number): void {
		this.currentTime = seconds;
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

	#startPlayback(source: string, resumeSeconds: number): void {
		this.source = source;
		this.currentTime = resumeSeconds;
		this.paused = false;
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
