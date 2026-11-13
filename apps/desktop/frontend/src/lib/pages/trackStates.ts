import type { Track } from '#lib/api/music.ts';
import { downloads } from '#lib/downloads/downloads.svelte.ts';
import { library } from '#lib/library/library.svelte.ts';

/** Looks up ratings and download states of tracks that are shown in rows, so the rows can draw them. */
export function ensureTrackStates(tracks: Track[]): void {
	const videoIds = tracks.map((track) => track.videoId);
	library.ensureRatings(videoIds);
	downloads.ensureStates(videoIds);
}
