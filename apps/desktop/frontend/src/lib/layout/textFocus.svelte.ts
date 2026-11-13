/**
 * Whether a text field has the keyboard focus, so that shortcuts such as Space for play and pause
 * leave the typing alone. The host does not tell the app which element is focused, so every
 * `Input` reports its focus here: `onfocus={() => textFocus.enter()}` and `onblur={() => textFocus.leave()}`.
 */
class TextFocus {
	#focusedFieldCount = $state(0);

	typing = $derived(this.#focusedFieldCount > 0);

	enter(): void {
		this.#focusedFieldCount += 1;
	}

	leave(): void {
		this.#focusedFieldCount = Math.max(0, this.#focusedFieldCount - 1);
	}
}

export const textFocus = new TextFocus();
