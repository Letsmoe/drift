import {
	beginNeoworksSignIn,
	cancelNeoworksSignIn,
	loadNeoworksStatus,
	onNeoworksChange,
	signOutOfNeoworks,
	syncNeoworks,
	type NeoworksAccountStatus
} from '#lib/api/neoworks.ts';
import { describeError } from '#lib/format.ts';

/** The optional Neoworks account that syncs the library across devices. */
export class NeoworksAccount {
	status = $state<NeoworksAccountStatus | null>(null);
	/** A failure of an action started here; sign-in and sync failures come in the status. */
	actionError = $state('');

	#started = false;

	get available(): boolean {
		if (!this.status) {
			return false;
		}
		return this.status.available;
	}

	get signedIn(): boolean {
		if (!this.status) {
			return false;
		}
		return this.status.signedIn;
	}

	get signingIn(): boolean {
		if (!this.status) {
			return false;
		}
		return this.status.signingIn;
	}

	/** The name to show: the account's name, else its email. */
	get displayName(): string {
		if (!this.status) {
			return '';
		}
		if (this.status.name !== '') {
			return this.status.name;
		}
		return this.status.email;
	}

	get errorMessage(): string {
		if (this.actionError !== '') {
			return this.actionError;
		}
		if (!this.status) {
			return '';
		}
		return this.status.error;
	}

	/** Subscribes to the backend's status; runs once. */
	start(): void {
		if (this.#started) {
			return;
		}
		this.#started = true;
		onNeoworksChange((status) => (this.status = status));
		loadNeoworksStatus().then(
			(status) => (this.status = status),
			(error) => (this.actionError = describeError(error))
		);
	}

	/** Opens the Neoworks sign-in in the browser; the backend takes the callback. */
	async signIn(): Promise<void> {
		this.actionError = '';
		try {
			const address = await beginNeoworksSignIn();
			window.open(address, '_blank');
		} catch (error) {
			this.actionError = describeError(error);
		}
	}

	cancelSignIn(): Promise<void> {
		return this.#run(() => cancelNeoworksSignIn());
	}

	signOut(): Promise<void> {
		return this.#run(() => signOutOfNeoworks());
	}

	sync(): Promise<void> {
		return this.#run(() => syncNeoworks());
	}

	async #run(action: () => Promise<void>): Promise<void> {
		this.actionError = '';
		try {
			await action();
		} catch (error) {
			this.actionError = describeError(error);
		}
	}
}

export const neoworksAccount = new NeoworksAccount();
