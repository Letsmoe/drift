import type { Track } from '#lib/api/music.ts';
import type { MenuEntry } from '#lib/menu/menuEntry.ts';
import { player } from '#lib/player/player.svelte.ts';
import ListPlusIcon from 'phosphor-svelte/lib/ListPlusIcon';
import QueueIcon from 'phosphor-svelte/lib/QueueIcon';

/** Menu for the "..." button of an album, playlist or artist page. */
export function collectionMenuEntries(tracks: Track[]): MenuEntry[] {
	return [
		{ label: 'Add to queue', icon: QueueIcon, onselect: () => player.addToQueue(tracks) },
		{ label: 'Play next', icon: ListPlusIcon, onselect: () => player.playNext(tracks) }
	];
}
