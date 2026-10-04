import { goto } from '$app/navigation';
import type { Artist, ItemKind, Track } from '#lib/api/music.ts';
import { downloads } from '#lib/downloads/downloads.svelte.ts';
import { itemHref } from '#lib/format.ts';
import { library } from '#lib/library/library.svelte.ts';
import type { MenuEntry } from '#lib/menu/menuEntry.ts';
import { player } from '#lib/player/player.svelte.ts';
import ArrowCircleDownIcon from 'phosphor-svelte/lib/ArrowCircleDownIcon';
import DiscIcon from 'phosphor-svelte/lib/DiscIcon';
import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
import ListPlusIcon from 'phosphor-svelte/lib/ListPlusIcon';
import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
import PlaylistIcon from 'phosphor-svelte/lib/PlaylistIcon';
import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
import QueueIcon from 'phosphor-svelte/lib/QueueIcon';
import TrashIcon from 'phosphor-svelte/lib/TrashIcon';
import UserIcon from 'phosphor-svelte/lib/UserIcon';

export interface TrackMenuOptions {
	kind?: ItemKind;
	/** Set on a user playlist page to offer "Remove from this playlist". */
	onRemoveFromPlaylist?: () => void;
}

export function buildTrackMenuEntries(track: Track, options: TrackMenuOptions = {}): MenuEntry[] {
	const kind = menuKind(options);
	const entries: MenuEntry[] = [
		{ label: 'Add to playlist', icon: PlaylistIcon, children: playlistEntries(track) },
		{ label: 'Add to queue', icon: QueueIcon, onselect: () => player.addToQueue([track]) },
		{ label: 'Play next', icon: ListPlusIcon, onselect: () => player.playNext([track]) },
		...navigationEntries(track),
		likeEntry(track, kind),
		downloadEntry(track, kind),
		{
			label: 'Not interested',
			icon: ProhibitIcon,
			onselect: () => library.dislike(track, kind)
		}
	];
	if (options.onRemoveFromPlaylist) {
		entries.push({
			label: 'Remove from this playlist',
			icon: TrashIcon,
			onselect: options.onRemoveFromPlaylist
		});
	}
	return entries;
}

function menuKind(options: TrackMenuOptions): ItemKind {
	if (options.kind) {
		return options.kind;
	}
	return 'song';
}

function playlistEntries(track: Track): MenuEntry[] {
	const entries: MenuEntry[] = [
		{ label: 'New playlist', icon: PlusIcon, onselect: () => addToNewPlaylist(track) }
	];
	for (const summary of library.playlists) {
		entries.push({
			label: summary.name,
			onselect: () => library.addToPlaylist(summary.id, [track])
		});
	}
	return entries;
}

async function addToNewPlaylist(track: Track): Promise<void> {
	const created = await library.createNumberedPlaylist();
	await library.addToPlaylist(created.id, [track]);
}

function navigationEntries(track: Track): MenuEntry[] {
	const entries: MenuEntry[] = [];
	const artists = track.artists.filter((artist) => artist.id !== '');
	if (artists.length > 0) {
		entries.push(artistEntry(artists));
	}
	if (track.album && track.album.id !== '') {
		const albumHref = itemHref('album', track.album.id);
		entries.push({
			label: 'Go to album',
			icon: DiscIcon,
			onselect: () => goto(albumHref),
			separatorBefore: artists.length === 0
		});
	}
	return entries;
}

function artistEntry(artists: Artist[]): MenuEntry {
	if (artists.length === 1) {
		return goToArtistEntry('Go to artist', artists[0], true);
	}
	return {
		label: 'Go to artist',
		icon: UserIcon,
		separatorBefore: true,
		children: artists.map((artist) => goToArtistEntry(artist.name, artist, false))
	};
}

function goToArtistEntry(label: string, artist: Artist, separatorBefore: boolean): MenuEntry {
	return {
		label,
		icon: UserIcon,
		separatorBefore,
		onselect: () => goto(itemHref('artist', artist.id))
	};
}

function likeEntry(track: Track, kind: ItemKind): MenuEntry {
	const liked = library.isLiked(track.videoId);
	let label = 'Save to Liked Songs';
	if (liked) {
		label = 'Remove from Liked Songs';
	}
	return {
		label,
		icon: HeartIcon,
		separatorBefore: true,
		onselect: () => library.toggleLike(track, kind)
	};
}

function downloadEntry(track: Track, kind: ItemKind): MenuEntry {
	if (downloads.stateOf(track.videoId) === '') {
		return {
			label: 'Download',
			icon: ArrowCircleDownIcon,
			onselect: () => downloads.download(track, kind)
		};
	}
	return {
		label: 'Remove download',
		icon: ArrowCircleDownIcon,
		onselect: () => downloads.remove(track.videoId)
	};
}
