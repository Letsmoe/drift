// The one toast of the window: a short message about something that just happened, such as a song added to
// a playlist. `App.svelte` shows it with the kit's Snackbar.

class Toast {
	message = $state('');
	open = $state(false);
	/** Bumped by every message, so a new one restarts the countdown of one still showing. */
	version = $state(0);

	show(message: string): void {
		this.message = message;
		this.version += 1;
		this.open = true;
	}
}

export const toast = new Toast();
