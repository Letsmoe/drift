import { invoke, listen } from '@neoworks/svelte-gpui';
import { untrack } from 'svelte';
import type { Track } from '#lib/api/music.ts';
import type { MprisRequest, MprisTrack, MprisUpdate } from '#lib/backend/commands.ts';
import type { Player } from '#lib/player/player.svelte.ts';

/**
 * Shows the player to MPRIS (the shell's apps/desktop/shell/src/mpris.rs), so media keys, desktop
 * widgets and `playerctl` can see and steer it. The player's state goes out as `mpris_update`
 * whenever the song, play state, volume or what Next and Previous can do changes, and after seeks;
 * requests come back as events. Call from a component's script: it registers effects.
 */
export function connectMpris(player: Player): void {
	$effect(() => listen<MprisRequest>('mpris', (request) => handleRequest(player, request)));
	$effect(() => pushWhenStateChanges(player));
	$effect(() => player.onSeek((seconds) => pushAfterSeek(player, seconds)));
}

function pushWhenStateChanges(player: Player): void {
	const state = readState(player);
	// The position is read without tracking it: MPRIS counts on from it, so it is not re-sent every 250 ms.
	const positionSeconds = untrack(() => player.currentTime);
	send({ ...state, positionSeconds, seeked: false });
}

function pushAfterSeek(player: Player, positionSeconds: number): void {
	untrack(() => send({ ...readState(player), positionSeconds, seeked: true }));
}

function send(update: MprisUpdate): void {
	invoke('mpris_update', update).catch((error: unknown) =>
		console.warn('mpris_update failed', error)
	);
}

function readState(player: Player): Omit<MprisUpdate, 'positionSeconds' | 'seeked'> {
	return {
		track: mprisTrackOf(player),
		status: statusOf(player),
		volume: volumeOf(player),
		canGoNext: canGoNext(player),
		canGoPrevious: player.track !== null
	};
}

function mprisTrackOf(player: Player): MprisTrack | null {
	const track = player.track;
	const entry = player.entry;
	if (track === null || entry === undefined) {
		return null;
	}
	let album = '';
	if (track.album) {
		album = track.album.name;
	}
	return {
		trackId: String(entry.id),
		title: track.title,
		artists: track.artists.map((artist) => artist.name),
		album,
		coverUrl: largestCoverUrl(track),
		lengthSeconds: player.duration
	};
}

function largestCoverUrl(track: Track): string {
	let url = '';
	let largestWidth = -1;
	for (const thumbnail of track.thumbnails) {
		if (thumbnail.width > largestWidth) {
			url = thumbnail.url;
			largestWidth = thumbnail.width;
		}
	}
	return url;
}

function statusOf(player: Player): MprisUpdate['status'] {
	if (player.track === null) {
		return 'stopped';
	}
	if (player.paused) {
		return 'paused';
	}
	return 'playing';
}

function volumeOf(player: Player): number {
	if (player.muted) {
		return 0;
	}
	return player.volume;
}

function canGoNext(player: Player): boolean {
	const queue = player.queue;
	if (queue.entries.length === 0) {
		return false;
	}
	return queue.index < queue.entries.length - 1 || queue.repeat !== 'off' || queue.autoplay;
}

function handleRequest(player: Player, request: MprisRequest): void {
	switch (request.kind) {
		case 'play':
			playIfPaused(player, true);
			break;
		case 'pause':
		case 'stop':
			playIfPaused(player, false);
			break;
		case 'playPause':
			player.toggle();
			break;
		case 'next':
			void player.next();
			break;
		case 'previous':
			void player.previous();
			break;
		case 'seek':
			player.seek(clampPosition(player, player.currentTime + request.offsetSeconds));
			break;
		case 'setPosition':
			player.seek(clampPosition(player, request.positionSeconds));
			break;
		case 'setVolume':
			player.setVolume(Math.min(Math.max(request.volume, 0), 1));
			break;
	}
}

function playIfPaused(player: Player, shouldPlay: boolean): void {
	if (player.paused === shouldPlay) {
		player.toggle();
	}
}

function clampPosition(player: Player, seconds: number): number {
	const upperBound = Math.max(player.duration, 0);
	return Math.min(Math.max(seconds, 0), upperBound);
}
