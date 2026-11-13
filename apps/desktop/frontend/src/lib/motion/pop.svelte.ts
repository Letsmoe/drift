// A pop: an element grows for a moment and springs back to its size, to acknowledge an action such as a
// like. The element needs a transition whose scale springs (`hoverTransition` from `#lib/theme.ts`); the
// host then eases its `scale` there and back.

/** How large the element grows. */
const popScale = 1.3;

/** How long it stays large before springing back: just long enough for the grow to show. */
const popHoldMs = 130;

export class Pop {
	#popping = $state(false);
	#timer: ReturnType<typeof setTimeout> | undefined = undefined;

	/** Grows the element, and lets it spring back after a moment. */
	trigger(): void {
		this.#popping = true;
		clearTimeout(this.#timer);
		this.#timer = setTimeout(() => {
			this.#popping = false;
		}, popHoldMs);
	}

	/** The scale to put in the element's style. */
	scale(): number {
		if (this.#popping) {
			return popScale;
		}
		return 1;
	}
}
