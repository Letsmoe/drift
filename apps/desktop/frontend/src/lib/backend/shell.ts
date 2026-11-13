import { files, invoke } from '@neoworks/svelte-gpui';

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

let address: BackendAddress | undefined;

/**
 * Asks the Rust shell where the backend listens. Without the shell (the plain svelte-gpui host)
 * the address baked in at build time is used, so the UI can run against a backend started by hand.
 * Call once before the app mounts; `backendAddress()` answers from then on.
 */
export async function loadBackendAddress(): Promise<void> {
	try {
		address = await invoke('backend_address');
		return;
	} catch (error) {
		console.warn('no Rust shell, using the build-time backend address:', error);
	}
	address = { url: __DRIFT_BACKEND_URL__, token: __DRIFT_BACKEND_TOKEN__ };
}

export function backendAddress(): BackendAddress {
	if (!address) {
		throw new Error('The backend address is not loaded yet.');
	}
	return address;
}

/** Opens an http(s) link in the default browser. */
export async function openExternal(url: string): Promise<void> {
	try {
		await invoke('open_external', { url });
	} catch (error) {
		console.warn('could not open the link', error);
	}
}

/** Resolves to the chosen path, or '' when the user cancels. */
export async function pickFile(options: OpenDialogOptions): Promise<string> {
	const picked = await files.pickFiles({ prompt: options.title });
	return firstPath(picked);
}

export async function pickFolder(options: OpenDialogOptions): Promise<string> {
	const picked = await files.pickFiles({ directories: true, prompt: options.title });
	return firstPath(picked);
}

function firstPath(picked: string[] | null): string {
	if (!picked || picked.length === 0) {
		return '';
	}
	return picked[0];
}
