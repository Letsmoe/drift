import {
	detectBrowserSessions,
	loadAccountStatus,
	signInFromBrowser,
	signInWithCookies,
	signOut,
	type AccountStatus,
	type BrowserSession
} from '#lib/api/youtubeAccount.ts';
import { describeError } from '#lib/format.ts';

/** The YouTube Music sign-in: the stored account, browser sessions found on this computer, pasted cookies. */
export class YouTubeAccountState {
	status = $state<AccountStatus | null>(null);
	browserSessions = $state<BrowserSession[]>([]);
	/** True once a browser search ran, so an empty list can say "none found". */
	searched = $state(false);
	busy = $state(false);
	pastedCookies = $state('');
	errorMessage = $state('');

	get signedIn(): boolean {
		return this.status?.signedIn === true;
	}

	get usableSessions(): BrowserSession[] {
		return this.browserSessions.filter((session) => session.hasYouTubeSession);
	}

	get unusableSessions(): BrowserSession[] {
		return this.browserSessions.filter((session) => !session.hasYouTubeSession);
	}

	/** True when the last error says the stored session no longer works. */
	get sessionExpired(): boolean {
		return this.errorMessage.includes('sign in again');
	}

	async refresh(): Promise<void> {
		try {
			this.status = await loadAccountStatus();
		} catch (error) {
			this.errorMessage = describeError(error);
		}
	}

	/** Searches the browsers on this computer; call it only from a click. */
	async searchBrowsers(): Promise<void> {
		this.errorMessage = '';
		this.busy = true;
		try {
			this.browserSessions = await detectBrowserSessions();
			this.searched = true;
		} catch (error) {
			this.errorMessage = describeError(error);
		} finally {
			this.busy = false;
		}
	}

	/** Returns true when the sign-in worked. */
	async useBrowserSession(session: BrowserSession): Promise<boolean> {
		return this.#signIn(() => signInFromBrowser(session.browser, session.profile));
	}

	/** Returns true when the sign-in worked. */
	async usePastedCookies(): Promise<boolean> {
		const pasted = this.pastedCookies;
		const worked = await this.#signIn(() => signInWithCookies(pasted));
		if (worked) {
			this.pastedCookies = '';
		}
		return worked;
	}

	async signOut(): Promise<void> {
		this.errorMessage = '';
		try {
			await signOut();
			this.status = await loadAccountStatus();
		} catch (error) {
			this.errorMessage = describeError(error);
		}
	}

	clearError(): void {
		this.errorMessage = '';
	}

	async #signIn(signInRequest: () => Promise<AccountStatus>): Promise<boolean> {
		this.errorMessage = '';
		this.busy = true;
		try {
			this.status = await signInRequest();
			return true;
		} catch (error) {
			this.errorMessage = describeError(error);
			return false;
		} finally {
			this.busy = false;
		}
	}
}
