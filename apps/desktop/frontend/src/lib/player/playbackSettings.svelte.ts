const crossfadeKey = 'drift:crossfade';
const crossfadeSecondsKey = 'drift:crossfade-seconds';

export const crossfadeSecondsLimits = { min: 1, max: 12 };
const defaultCrossfadeSeconds = 6;

function loadCrossfade(): boolean {
	try {
		return localStorage.getItem(crossfadeKey) === 'true';
	} catch {
		return false;
	}
}

function loadCrossfadeSeconds(): number {
	let stored: number;
	try {
		stored = Number(localStorage.getItem(crossfadeSecondsKey));
	} catch {
		return defaultCrossfadeSeconds;
	}
	if (!Number.isFinite(stored) || stored < crossfadeSecondsLimits.min) {
		return defaultCrossfadeSeconds;
	}
	return Math.min(stored, crossfadeSecondsLimits.max);
}

/** How songs follow each other. Crossfade is off by default. */
class PlaybackSettings {
	crossfade = $state(loadCrossfade());
	crossfadeSeconds = $state(loadCrossfadeSeconds());

	setCrossfade(enabled: boolean): void {
		this.crossfade = enabled;
		localStorage.setItem(crossfadeKey, String(enabled));
	}

	setCrossfadeSeconds(seconds: number): void {
		this.crossfadeSeconds = seconds;
		localStorage.setItem(crossfadeSecondsKey, String(seconds));
	}
}

export const playbackSettings = new PlaybackSettings();
