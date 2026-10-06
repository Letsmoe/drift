import {
	DetectBrowserSessions,
	YouTubeAccountStatus,
	YouTubeImport,
	YouTubePreview,
	YouTubeSignInFromBrowser,
	YouTubeSignInWithCookies,
	YouTubeSignOut
} from '#lib/backend/generated/App.ts';
import type { account, browsercookies } from '#lib/backend/generated/models.ts';
import type { ImportOptions, ImportSummary } from '#lib/api/musicImport.ts';

export type AccountStatus = account.Status;
export type BrowserSession = browsercookies.BrowserSession;

/** Who is signed in to YouTube Music and when the library was last synced; reads local state only. */
export function loadAccountStatus(): Promise<AccountStatus> {
	return YouTubeAccountStatus();
}

/** Looks through the browsers on this computer; opens their cookie stores, so call it on a click. */
export function detectBrowserSessions(): Promise<BrowserSession[]> {
	return DetectBrowserSessions();
}

export function signInFromBrowser(browser: string, profile: string): Promise<AccountStatus> {
	return YouTubeSignInFromBrowser(browser, profile);
}

/** Signs in with pasted text: a Cookie header, request headers or a "Copy as cURL" command. */
export function signInWithCookies(pastedText: string): Promise<AccountStatus> {
	return YouTubeSignInWithCookies(pastedText);
}

export function signOut(): Promise<void> {
	return YouTubeSignOut();
}

export function previewAccount(): Promise<ImportSummary> {
	return YouTubePreview();
}

export function startAccountImport(options: ImportOptions): Promise<void> {
	return YouTubeImport(options);
}
