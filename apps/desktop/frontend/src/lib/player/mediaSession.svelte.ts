import type { Track } from '#lib/api/music.ts';
import { formatArtists } from '#lib/format.ts';
import type { Player } from '#lib/player/player.svelte.ts';

/**
 * Hands the player to the Media Session, so the keyboard's media keys and the system's media
 * controls (MPRIS on Linux) can play, pause and skip. Chromium only handles play and pause on
 * its own; next and previous need these handlers. Call from a component's script.
 */
export function connectMediaSession(player: Player): void {
	if (!('mediaSession' in navigator)) {
		return;
	}
	setActionHandlers(player);
	$effect(() => {
		navigator.mediaSession.metadata = metadataOf(player.track);
	});
	$effect(() => {
		navigator.mediaSession.playbackState = playbackStateOf(player);
	});
}

function setActionHandlers(player: Player): void {
	const session = navigator.mediaSession;
	session.setActionHandler('play', () => playIfPaused(player, true));
	session.setActionHandler('pause', () => playIfPaused(player, false));
	session.setActionHandler('nexttrack', () => void player.next());
	session.setActionHandler('previoustrack', () => void player.previous());
	session.setActionHandler('seekto', (details) => {
		if (details.seekTime !== undefined) {
			player.seek(details.seekTime);
		}
	});
}

function playIfPaused(player: Player, shouldPlay: boolean): void {
	if (player.paused === shouldPlay) {
		player.toggle();
	}
}

function metadataOf(track: Track | null): MediaMetadata | null {
	if (!track) {
		return null;
	}
	const artwork = track.thumbnails.map((thumbnail) => ({
		src: thumbnail.url,
		sizes: `${thumbnail.width}x${thumbnail.height}`
	}));
	let album = '';
	if (track.album) {
		album = track.album.name;
	}
	return new MediaMetadata({
		title: track.title,
		artist: formatArtists(track.artists),
		album,
		artwork
	});
}

function playbackStateOf(player: Player): MediaSessionPlaybackState {
	if (!player.track) {
		return 'none';
	}
	if (player.paused) {
		return 'paused';
	}
	return 'playing';
}
