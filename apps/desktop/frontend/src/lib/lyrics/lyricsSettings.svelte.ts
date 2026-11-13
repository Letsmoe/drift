import { getItem, setItem } from '#lib/storage.ts';

const storageKey = 'drift:lyrics-allow-lrclib';

function loadAllowLRCLIB(): boolean {
	return getItem(storageKey) === 'true';
}

/** Whether lyrics may be looked up on lrclib.net when YouTube has no synced lyrics. Off by default. */
class LyricsSettings {
	allowLRCLIB = $state(loadAllowLRCLIB());

	setAllowLRCLIB(allow: boolean): void {
		this.allowLRCLIB = allow;
		setItem(storageKey, String(allow));
	}
}

export const lyricsSettings = new LyricsSettings();
