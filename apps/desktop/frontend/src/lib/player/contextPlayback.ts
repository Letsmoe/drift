import type { ItemKind, Track } from '#lib/api/music.ts';
import { player } from '#lib/player/player.svelte.ts';

/** True when the queue was started from the given context (`album:<id>`, `liked`, ...). */
export function isContextCurrent(context: string): boolean {
	return player.queue.context === context && player.track !== null;
}

export function isContextPlaying(context: string): boolean {
	return isContextCurrent(context) && !player.paused;
}

/** Starts the context from the top (or a random track with shuffle on), or toggles it when it is current. */
export function playContext(tracks: Track[], context: string, kind: ItemKind = 'song'): void {
	if (tracks.length === 0) {
		return;
	}
	if (isContextCurrent(context)) {
		player.toggle();
		return;
	}
	player.playTracks(tracks, firstIndex(tracks.length), context, kind);
}

function firstIndex(trackCount: number): number {
	if (!player.queue.shuffle) {
		return 0;
	}
	return Math.floor(Math.random() * trackCount);
}
