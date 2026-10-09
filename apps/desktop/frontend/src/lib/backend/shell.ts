/** Where the Go backend listens and the token its RPC socket requires. */
export interface BackendAddress {
	url: string;
	token: string;
}

export interface FileFilter {
	name: string;
	extensions: string[];
}

export interface OpenDialogOptions {
	title: string;
	filters?: FileFilter[];
}

/** What electron/src/preload.ts exposes as `window.drift`. */
interface ShellBridge {
	backend(): BackendAddress;
	/** Resolves to the chosen path, or '' when the user cancels. */
	pickFile(options: OpenDialogOptions): Promise<string>;
	pickFolder(options: OpenDialogOptions): Promise<string>;
}

declare global {
	interface Window {
		drift?: ShellBridge;
	}
}

/**
 * The backend address from the Electron shell. Opened in a plain browser, the page
 * is served by the backend itself, so its origin and `?token=` are used instead.
 */
export function backendAddress(): BackendAddress {
	if (window.drift) {
		return window.drift.backend();
	}
	const token = new URLSearchParams(window.location.search).get('token');
	return { url: window.location.origin, token: token || '' };
}

function requireShell(): ShellBridge {
	if (!window.drift) {
		throw new Error('File dialogs need the desktop app.');
	}
	return window.drift;
}

export function pickFile(options: OpenDialogOptions): Promise<string> {
	return requireShell().pickFile(options);
}

export function pickFolder(options: OpenDialogOptions): Promise<string> {
	return requireShell().pickFolder(options);
}
