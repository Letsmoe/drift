import { getItem, setItem } from '#lib/storage.ts';

const crossfadeKey = 'drift:crossfade';
const crossfadeSecondsKey = 'drift:crossfade-seconds';

export const crossfadeSecondsLimits = { min: 1, max: 12 };
const defaultCrossfadeSeconds = 6;

function loadCrossfade(): boolean {
	return getItem(crossfadeKey) === 'true';
}

function loadCrossfadeSeconds(): number {
	const stored = Number(getItem(crossfadeSecondsKey));
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
		setItem(crossfadeKey, String(enabled));
	}

	setCrossfadeSeconds(seconds: number): void {
		this.crossfadeSeconds = seconds;
		setItem(crossfadeSecondsKey, String(seconds));
	}
}

export const playbackSettings = new PlaybackSettings();
