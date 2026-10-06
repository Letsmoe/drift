import { afterNavigate } from '$app/navigation';

// SvelteKit 3 doesn't export the callback's argument type; derive it from afterNavigate.
type AfterNavigate = Parameters<Parameters<typeof afterNavigate>[0]>[0];

/**
 * Back/forward availability of the app's own history, counted from navigations:
 * position 0 is the page the app opened on. Create it during component
 * initialization: it registers an afterNavigate callback.
 */
export class NavigationHistory {
	#position = $state(0);
	#lastPosition = $state(0);

	canGoBack = $derived(this.#position > 0);
	canGoForward = $derived(this.#position < this.#lastPosition);

	constructor() {
		afterNavigate((navigation) => this.#record(navigation));
	}

	back(): void {
		if (this.canGoBack) {
			window.history.back();
		}
	}

	forward(): void {
		if (this.canGoForward) {
			window.history.forward();
		}
	}

	#record(navigation: AfterNavigate): void {
		if (navigation.type === 'enter') {
			return;
		}
		if (navigation.type === 'popstate') {
			this.#position = clamp(this.#position + navigation.delta, 0, this.#lastPosition);
			return;
		}
		// A new entry drops everything that was ahead of it.
		this.#position += 1;
		this.#lastPosition = this.#position;
	}
}

function clamp(value: number, lowest: number, highest: number): number {
	return Math.min(Math.max(value, lowest), highest);
}
