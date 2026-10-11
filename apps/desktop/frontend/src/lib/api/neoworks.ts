import {
	NeoworksCancelSignIn,
	NeoworksSignIn,
	NeoworksSignOut,
	NeoworksStatus,
	NeoworksSync
} from '#lib/backend/generated/App.ts';
import { onEvent } from '#lib/backend/rpc.ts';
import type { main } from '#lib/backend/generated/models.ts';

export type NeoworksAccountStatus = main.NeoworksStatus;

export function loadNeoworksStatus(): Promise<NeoworksAccountStatus> {
	return NeoworksStatus();
}

/** Starts signing in; resolves to the Neoworks URL to open in the browser. */
export function beginNeoworksSignIn(): Promise<string> {
	return NeoworksSignIn();
}

export function cancelNeoworksSignIn(): Promise<void> {
	return NeoworksCancelSignIn();
}

/** Signs out and wipes the synced copy; the local library stays. */
export function signOutOfNeoworks(): Promise<void> {
	return NeoworksSignOut();
}

/** Syncs now; the outcome arrives through onNeoworksChange. */
export function syncNeoworks(): Promise<void> {
	return NeoworksSync();
}

export function onNeoworksChange(listener: (status: NeoworksAccountStatus) => void): () => void {
	return onEvent<NeoworksAccountStatus>('neoworks', listener);
}
