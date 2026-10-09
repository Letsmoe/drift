const storageKey = 'drift:lyrics-allow-lrclib';

function loadAllowLRCLIB(): boolean {
	try {
		return localStorage.getItem(storageKey) === 'true';
	} catch {
		return false;
	}
}

/** Whether lyrics may be looked up on lrclib.net when YouTube has no synced lyrics. Off by default. */
class LyricsSettings {
	allowLRCLIB = $state(loadAllowLRCLIB());

	setAllowLRCLIB(allow: boolean): void {
		this.allowLRCLIB = allow;
		localStorage.setItem(storageKey, String(allow));
	}
}

export const lyricsSettings = new LyricsSettings();
