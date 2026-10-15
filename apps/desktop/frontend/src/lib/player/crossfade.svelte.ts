import type { Player } from '#lib/player/player.svelte.ts';
import { playbackSettings } from '#lib/player/playbackSettings.svelte.ts';

const fadeStepMs = 50;
/** Songs shorter than this many crossfade lengths end without one, so they don't fade most of their length. */
const minimumLengthInFades = 3;
/** Longest wait for the second element to take over the ending song before giving up on the crossfade. */
const handOverTimeoutMs = 3000;

/**
 * Crossfades into the next song. The player's own <audio> element always plays the current
 * song; near its end a second element takes over the ending song at the same position and
 * fades it out, while the main element moves on to the next song and fades it in.
 */
export class Crossfade {
	/** Multiplier on the main element's volume while the next song fades in. */
	incomingGain = $state(1);

	#player: Player;
	#main: HTMLAudioElement | undefined = undefined;
	#outgoing: HTMLAudioElement | undefined = undefined;
	#starting = false;
	#fadeTimer: ReturnType<typeof setInterval> | undefined = undefined;
	#fadeStartedAt = 0;
	#incomingEntryId: number | undefined = undefined;

	constructor(player: Player) {
		this.#player = player;
		// Playing, pausing or seeking by the user ends the crossfade.
		player.onUserAction(() => this.finish());
	}

	attach(main: HTMLAudioElement, outgoing: HTMLAudioElement): void {
		this.#main = main;
		this.#outgoing = outgoing;
	}

	/** Called as the song plays: starts the crossfade once the song is close enough to its end. */
	check(): void {
		if (this.#shouldStart()) {
			void this.#start();
		}
	}

	/** Ends a running crossfade at once: the ending song stops, the next one plays at full volume. */
	finish(): void {
		if (this.#fadeTimer !== undefined) {
			clearInterval(this.#fadeTimer);
			this.#fadeTimer = undefined;
		}
		this.#incomingEntryId = undefined;
		this.incomingGain = 1;
		this.#stopOutgoing();
	}

	#shouldStart(): boolean {
		const main = this.#main;
		if (!main || !playbackSettings.crossfade || this.#starting || this.#fadeTimer !== undefined) {
			return false;
		}
		if (main.paused || !Number.isFinite(main.duration)) {
			return false;
		}
		const fadeSeconds = playbackSettings.crossfadeSeconds;
		if (main.duration < fadeSeconds * minimumLengthInFades) {
			return false;
		}
		return main.duration - main.currentTime <= fadeSeconds && this.#hasNextEntry();
	}

	#hasNextEntry(): boolean {
		const queue = this.#player.queue;
		if (queue.repeat === 'one') {
			return false;
		}
		return queue.index < queue.entries.length - 1 || queue.repeat === 'all';
	}

	async #start(): Promise<void> {
		this.#starting = true;
		const handedOver = await this.#handOver();
		this.#starting = false;
		if (!handedOver) {
			this.#stopOutgoing();
			return;
		}
		this.incomingGain = 0;
		this.#fadeStartedAt = performance.now();
		this.#fadeTimer = setInterval(() => this.#step(), fadeStepMs);
		await this.#player.trackEnded();
		this.#incomingEntryId = this.#player.entry?.id;
	}

	// The second element plays the ending song muted until it runs in step with the main one.
	async #handOver(): Promise<boolean> {
		const main = this.#main;
		const outgoing = this.#outgoing;
		if (!main || !outgoing) {
			return false;
		}
		outgoing.muted = true;
		outgoing.src = main.currentSrc;
		outgoing.currentTime = main.currentTime;
		try {
			await withTimeout(outgoing.play(), handOverTimeoutMs);
			outgoing.currentTime = main.currentTime;
			await withTimeout(nextEvent(outgoing, 'seeked'), handOverTimeoutMs);
		} catch {
			return false;
		}
		if (main.paused) {
			return false;
		}
		outgoing.muted = this.#player.muted;
		outgoing.volume = this.#player.volume;
		return true;
	}

	#step(): void {
		if (this.#interrupted()) {
			this.finish();
			return;
		}
		const fadeMs = playbackSettings.crossfadeSeconds * 1000;
		const progress = Math.min((performance.now() - this.#fadeStartedAt) / fadeMs, 1);
		// Equal-power curves keep the loudness even while both songs play.
		this.incomingGain = Math.sin((progress * Math.PI) / 2);
		this.#setOutgoingVolume(Math.cos((progress * Math.PI) / 2));
		if (progress >= 1) {
			this.finish();
		}
	}

	// Picking another song ends the crossfade.
	#interrupted(): boolean {
		const entryId = this.#player.entry?.id;
		return this.#incomingEntryId !== undefined && entryId !== this.#incomingEntryId;
	}

	#setOutgoingVolume(gain: number): void {
		const outgoing = this.#outgoing;
		if (!outgoing) {
			return;
		}
		outgoing.muted = this.#player.muted;
		outgoing.volume = Math.max(0, Math.min(1, this.#player.volume * gain));
	}

	#stopOutgoing(): void {
		const outgoing = this.#outgoing;
		if (!outgoing || !outgoing.hasAttribute('src')) {
			return;
		}
		outgoing.pause();
		outgoing.removeAttribute('src');
		outgoing.load();
	}
}

function nextEvent(element: HTMLElement, type: string): Promise<void> {
	return new Promise((resolve) => element.addEventListener(type, () => resolve(), { once: true }));
}

function withTimeout<T>(promise: Promise<T>, timeoutMs: number): Promise<T> {
	return new Promise((resolve, reject) => {
		const timer = setTimeout(() => reject(new Error('timed out')), timeoutMs);
		promise.then(
			(value) => {
				clearTimeout(timer);
				resolve(value);
			},
			(error) => {
				clearTimeout(timer);
				reject(error);
			}
		);
	});
}
