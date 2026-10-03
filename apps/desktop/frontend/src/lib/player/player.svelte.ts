import { streamUrl, type Track } from '#lib/api/music.ts';
import { describeError } from '#lib/format.ts';

/**
 * Playback state shared by the track list and the player bar. The <audio> element
 * in PlayerBar binds to `source`, `paused`, `currentTime` and `duration`.
 */
export class Player {
	track = $state<Track | null>(null);
	source = $state<string | undefined>(undefined);
	paused = $state(true);
	currentTime = $state(0);
	duration = $state(0);
	buffering = $state(false);
	errorMessage = $state('');

	// A slow URL lookup must not override a track the user picked afterwards.
	#latestRequestId = 0;

	async play(track: Track): Promise<void> {
		this.#latestRequestId += 1;
		const requestId = this.#latestRequestId;

		this.track = track;
		this.errorMessage = '';
		this.currentTime = 0;
		this.duration = track.durationSeconds;
		this.buffering = true;

		try {
			const source = await streamUrl(track.videoId);
			if (requestId === this.#latestRequestId) {
				this.source = source;
				this.paused = false;
			}
		} catch (error) {
			if (requestId === this.#latestRequestId) {
				this.fail(describeError(error));
			}
		}
	}

	toggle(): void {
		if (!this.track || this.errorMessage) {
			return;
		}
		this.paused = !this.paused;
	}

	seek(seconds: number): void {
		this.currentTime = seconds;
	}

	isCurrent(track: Track): boolean {
		return this.track?.videoId === track.videoId;
	}

	fail(message: string): void {
		this.errorMessage = message;
		this.buffering = false;
		this.paused = true;
	}
}

export const player = new Player();
