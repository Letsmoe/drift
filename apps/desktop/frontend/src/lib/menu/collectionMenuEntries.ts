import type { Track } from '#lib/api/music.ts';
import { listPlus, queue } from '#lib/icons.ts';
import type { MenuEntry } from '#lib/menu/menuEntry.ts';
import { player } from '#lib/player/player.svelte.ts';

/** Menu for the "..." button of an album, playlist or artist page. */
export function collectionMenuEntries(tracks: Track[]): MenuEntry[] {
	return [
		{ label: 'Add to queue', icon: queue, onselect: () => player.addToQueue(tracks) },
		{ label: 'Play next', icon: listPlus, onselect: () => player.playNext(tracks) }
	];
}
