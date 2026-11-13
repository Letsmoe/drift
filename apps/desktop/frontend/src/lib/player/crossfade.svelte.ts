import type { AudioEvent } from '#lib/backend/commands.ts';
import {
	cancelAudioCrossfade,
	loadAudio,
	outgoingSlot,
	seekAudio,
	setAudioGain,
	startAudioCrossfade,
	stopAudio
} from '#lib/player/audioEngine.ts';
import type { Player } from '#lib/player/player.svelte.ts';
import { playbackSettings } from '#lib/player/playbackSettings.svelte.ts';

/** Songs shorter than this many crossfade lengths end without one, so they don't fade most of their length. */
const minimumLengthInFades = 3;
/** Longest wait for the second slot to take over the ending song before giving up on the crossfade. */
const handOverTimeoutMs = 3000;

/**
 * Decides when to crossfade into the next song. The audio engine's main slot always plays the
 * current song; near its end the second slot takes over the ending song at the same position, and
 * the engine then fades it out while the main slot, which moves on to the next song, fades in.
 * The fade itself runs in the engine (apps/desktop/shell/src/audio/fade.rs), not here.
 */
export class Crossfade {
	#player: Player;
	#starting = false;
	/** The engine is fading the two slots. */
	#fading = false;
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

	/** The engine says its crossfade is over, because it ran out or because something ended it. */
	handleFadeEnded(): void {
		this.#fading = false;
		this.#incomingEntryId = undefined;
		this.#outgoingPlaying = false;
	}

	/** Called when the queue shows another entry: picking another song ends the crossfade. */
	handleEntryChange(entryId: number | undefined): void {
		const incomingEntryId = this.#incomingEntryId;
		if (this.#fading && incomingEntryId !== undefined && entryId !== incomingEntryId) {
			this.finish();
		}
	}

	/** Ends a running crossfade at once: the ending song stops, the next one plays at full volume. */
	finish(): void {
		if (this.#fading) {
			cancelAudioCrossfade();
		}
		this.#fading = false;
		this.#incomingEntryId = undefined;
		this.#stopOutgoing();
	}

	#shouldStart(): boolean {
		const player = this.#player;
		if (!playbackSettings.crossfade || this.#starting || this.#fading) {
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
		this.#fading = true;
		startAudioCrossfade(playbackSettings.crossfadeSeconds);
		await this.#player.trackEnded();
		if (this.#fading) {
			this.#incomingEntryId = this.#player.entry?.id;
		}
	}

	// The second slot plays the ending song silently until it runs in step with the main one.
	async #handOver(): Promise<boolean> {
		const player = this.#player;
		const url = player.source;
		if (url === undefined) {
			return false;
		}
		setAudioGain(outgoingSlot, 0);
		const started = this.#nextOutgoingPosition();
		loadAudio(outgoingSlot, url, player.currentTime, true);
		this.#outgoingPlaying = true;
		if (!(await started)) {
			return false;
		}
		const aligned = this.#nextOutgoingPosition();
		seekAudio(outgoingSlot, player.currentTime);
		if (!(await aligned)) {
			return false;
		}
		// A play, pause or seek by the user in the meantime has already ended the hand-over.
		return this.#outgoingPlaying && !player.paused;
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

	#stopOutgoing(): void {
		if (!this.#outgoingPlaying) {
			return;
		}
		this.#outgoingPlaying = false;
		stopAudio(outgoingSlot);
	}
}
