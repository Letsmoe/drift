// A line too long for its box: it stays on one line, cut off at the box's edge, and while the pointer is over it
// glides to its end, rests, and glides back, until the pointer leaves. A line that fits never moves.

import { Tween } from 'svelte/motion';

/** How fast the line moves: slow enough to read while it passes. */
const pixelsPerSecond = 40;

/** How long the line rests at either end, and before it first moves. */
const restMs = 900;

/** How long a line that was moved takes to go back to its start when the pointer leaves. */
const returnMs = 250;

export class Marquee {
	#boxWidth = $state(0);
	#lineWidth = $state(0);
	#offset = new Tween(0, { duration: 0 });
	/** Bumped by every start and stop, so a glide that was overtaken does not carry on. */
	#run = 0;

	/** How much of the line does not fit in the box. */
	readonly overflow = $derived(Math.max(0, this.#lineWidth - this.#boxWidth));

	/** For the box's `onresize`. */
	measureBox(width: number): void {
		this.#boxWidth = width;
	}

	/** For the line's `onresize`. */
	measureLine(width: number): void {
		this.#lineWidth = width;
	}

	/** For the line's `translate`. */
	translate(): [number, number] {
		return [-this.#offset.current, 0];
	}

	/** The pointer is over the line: it glides back and forth while it does not fit. */
	start(): void {
		this.#run += 1;
		void this.#glideWhile(this.#run);
	}

	/** The pointer left: the line goes back to its start. */
	stop(): void {
		this.#run += 1;
		void this.#offset.set(0, { duration: returnMs });
	}

	async #glideWhile(run: number): Promise<void> {
		while (await this.#rest(run)) {
			if (this.overflow < 1) {
				return;
			}
			await this.#glideTo(this.overflow);
			if (!(await this.#rest(run))) {
				return;
			}
			await this.#glideTo(0);
		}
	}

	#glideTo(target: number): Promise<void> {
		const distance = Math.abs(target - this.#offset.current);
		return this.#offset.set(target, { duration: (distance / pixelsPerSecond) * 1000 });
	}

	/** Waits at an end; false once the pointer has left or another run began. */
	async #rest(run: number): Promise<boolean> {
		await new Promise((resolve) => setTimeout(resolve, restMs));
		return run === this.#run;
	}
}
