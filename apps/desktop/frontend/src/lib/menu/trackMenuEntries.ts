import { goto } from '$app/navigation';
import type { Artist, ItemKind, Track } from '#lib/api/music.ts';
import { downloads } from '#lib/downloads/downloads.svelte.ts';
import { itemHref } from '#lib/format.ts';
import { library } from '#lib/library/library.svelte.ts';
import type { MenuEntry } from '#lib/menu/menuEntry.ts';
import { player } from '#lib/player/player.svelte.ts';
import { recommendationFeedback } from '#lib/recommend/recommendationFeedback.svelte.ts';
import ArrowCircleDownIcon from 'phosphor-svelte/lib/ArrowCircleDownIcon';
import DiscIcon from 'phosphor-svelte/lib/DiscIcon';
import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
import ListPlusIcon from 'phosphor-svelte/lib/ListPlusIcon';
import MicrophoneIcon from 'phosphor-svelte/lib/MicrophoneIcon';
import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
import PlaylistIcon from 'phosphor-svelte/lib/PlaylistIcon';
import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
import ThumbsDownIcon from 'phosphor-svelte/lib/ThumbsDownIcon';
import UserMinusIcon from 'phosphor-svelte/lib/UserMinusIcon';
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
		...navigationEntries(track, kind),
		likeEntry(track, kind),
		downloadEntry(track, kind),
		...feedbackEntries(track, kind)
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

/** "episode" or "song", for labels like "Don't recommend this episode". */
function nounFor(kind: ItemKind): string {
	if (kind === 'episode') {
		return 'episode';
	}
	return 'song';
}

function feedbackEntries(track: Track, kind: ItemKind): MenuEntry[] {
	const entries: MenuEntry[] = [
		{
			label: dislikeLabel(track, kind),
			icon: ThumbsDownIcon,
			separatorBefore: true,
			onselect: () => recommendationFeedback.toggleDislike(track, kind)
		},
		{
			label: `Don’t recommend this ${nounFor(kind)}`,
			icon: ProhibitIcon,
			onselect: () => recommendationFeedback.blockTrack(track)
		}
	];
	if (kind === 'episode') {
		entries.push(...blockShowEntries(track));
		return entries;
	}
	const artists = track.artists.filter((artist) => artist.id !== '');
	if (artists.length === 1) {
		entries.push(blockArtistEntry(`Don’t recommend ${artists[0].name}`, artists[0]));
	} else if (artists.length > 1) {
		entries.push({
			label: 'Don’t recommend artist',
			icon: UserMinusIcon,
			children: artists.map((artist) => blockArtistEntry(artist.name, artist))
		});
	}
	return entries;
}

// An episode's "album" is its show.
function blockShowEntries(track: Track): MenuEntry[] {
	const show = track.album;
	if (!show || show.id === '') {
		return [];
	}
	return [
		{
			label: `Don’t recommend ${show.name}`,
			icon: UserMinusIcon,
			onselect: () => recommendationFeedback.blockPodcast(show.id, show.name)
		}
	];
}

function dislikeLabel(track: Track, kind: ItemKind): string {
	if (library.isDisliked(track.videoId)) {
		return 'Remove dislike';
	}
	return `I don’t like this ${nounFor(kind)}`;
}

function blockArtistEntry(label: string, artist: Artist): MenuEntry {
	return {
		label,
		icon: UserMinusIcon,
		onselect: () => recommendationFeedback.blockArtist(artist)
	};
}

/** "New playlist" plus one entry per local playlist, each adding track to it. */
// Newest playlist last: the menu opens near the bottom of the window, so the cursor lands there.
export function playlistEntries(track: Track): MenuEntry[] {
	const entries: MenuEntry[] = [
		{ label: 'New playlist', icon: PlusIcon, onselect: () => addToNewPlaylist(track) }
	];
	for (const summary of library.playlists.toReversed()) {
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

function navigationEntries(track: Track, kind: ItemKind): MenuEntry[] {
	const entries: MenuEntry[] = [];
	const artists = track.artists.filter((artist) => artist.id !== '');
	if (artists.length > 0) {
		entries.push(artistEntry(artists));
	}
	if (track.album && track.album.id !== '') {
		entries.push(albumEntry(track.album.id, kind, artists.length === 0));
	}
	return entries;
}

// An episode's "album" is its podcast show.
function albumEntry(albumId: string, kind: ItemKind, separatorBefore: boolean): MenuEntry {
	if (kind === 'episode') {
		return {
			label: 'Go to podcast',
			icon: MicrophoneIcon,
			onselect: () => goto(itemHref('podcast', albumId)),
			separatorBefore
		};
	}
	return {
		label: 'Go to album',
		icon: DiscIcon,
		onselect: () => goto(itemHref('album', albumId)),
		separatorBefore
	};
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
