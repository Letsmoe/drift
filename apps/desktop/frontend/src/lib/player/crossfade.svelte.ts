import type { AudioEvent } from '#lib/backend/commands.ts';
import {
	loadAudio,
	outgoingSlot,
	seekAudio,
	setAudioVolume,
	stopAudio
} from '#lib/player/audioEngine.ts';
import type { Player } from '#lib/player/player.svelte.ts';
import { playbackSettings } from '#lib/player/playbackSettings.svelte.ts';

const fadeStepMs = 50;
/** Songs shorter than this many crossfade lengths end without one, so they don't fade most of their length. */
const minimumLengthInFades = 3;
/** Longest wait for the second slot to take over the ending song before giving up on the crossfade. */
const handOverTimeoutMs = 3000;

/**
 * Crossfades into the next song. The audio engine's main slot always plays the current song; near
 * its end the second slot takes over the ending song at the same position and fades it out, while
 * the main slot moves on to the next song and fades it in.
 */
export class Crossfade {
	/** Multiplier on the main slot's volume while the next song fades in. */
	incomingGain = $state(1);

	#player: Player;
	#starting = false;
	#fadeTimer: ReturnType<typeof setInterval> | undefined = undefined;
	#fadeStartedAt = 0;
	#incomingEntryId: number | undefined = undefined;
	#outgoingPlaying = false;
	#positionWaiter: (() => void) | undefined = undefined;

	constructor(player: Player) {
		this.#player = player;
		// Playing, pausing or seeking by the user ends the crossfade.
		player.onUserAction(() => this.finish());
	}

	/** Called as the song plays: starts the crossfade once the song is close enough to its end. */
	check(): void {
		if (this.#shouldStart()) {
			void this.#start();
		}
	}

	/** Hears what the second slot reports while it takes over the ending song. */
	handleOutgoingEvent(event: AudioEvent): void {
		if (event.kind === 'position' && this.#positionWaiter) {
			this.#positionWaiter();
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
		const player = this.#player;
		if (!playbackSettings.crossfade || this.#starting || this.#fadeTimer !== undefined) {
			return false;
		}
		if (player.paused || player.source === undefined || !Number.isFinite(player.duration)) {
			return false;
		}
		const fadeSeconds = playbackSettings.crossfadeSeconds;
		if (player.duration < fadeSeconds * minimumLengthInFades) {
			return false;
		}
		return player.duration - player.currentTime <= fadeSeconds && this.#hasNextEntry();
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

	// The second slot plays the ending song silently until it runs in step with the main one.
	async #handOver(): Promise<boolean> {
		const player = this.#player;
		const url = player.source;
		if (url === undefined) {
			return false;
		}
		setAudioVolume(outgoingSlot, 0);
		const started = this.#nextOutgoingPosition();
		loadAudio(outgoingSlot, url, player.currentTime, true);
		this.#outgoingPlaying = true;
		if (!(await started)) {
			return false;
		}
		const aligned = this.#nextOutgoingPosition();
		seekAudio(outgoingSlot, player.currentTime);
		if (!(await aligned) || player.paused) {
			return false;
		}
		this.#setOutgoingGain(1);
		return true;
	}

	// Resolves true at the next position report of the second slot, false when none comes in time.
	#nextOutgoingPosition(): Promise<boolean> {
		return new Promise((resolve) => {
			const timer = setTimeout(() => {
				this.#positionWaiter = undefined;
				resolve(false);
			}, handOverTimeoutMs);
			this.#positionWaiter = () => {
				clearTimeout(timer);
				this.#positionWaiter = undefined;
				resolve(true);
			};
		});
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
		this.#setOutgoingGain(Math.cos((progress * Math.PI) / 2));
		if (progress >= 1) {
			this.finish();
		}
	}

	// Picking another song ends the crossfade.
	#interrupted(): boolean {
		const entryId = this.#player.entry?.id;
		return this.#incomingEntryId !== undefined && entryId !== this.#incomingEntryId;
	}

	#setOutgoingGain(gain: number): void {
		const player = this.#player;
		let volume = player.volume * gain;
		if (player.muted) {
			volume = 0;
		}
		setAudioVolume(outgoingSlot, Math.max(0, Math.min(1, volume)));
	}

	#stopOutgoing(): void {
		if (!this.#outgoingPlaying) {
			return;
		}
		this.#outgoingPlaying = false;
		stopAudio(outgoingSlot);
	}
}
