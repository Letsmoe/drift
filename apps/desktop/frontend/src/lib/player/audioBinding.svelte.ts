import { listen } from '@neoworks/svelte-gpui';
import { untrack } from 'svelte';
import type { AudioEvent } from '#lib/backend/commands.ts';
import {
	loadAudio,
	mainSlot,
	outgoingSlot,
	pauseAudio,
	playAudio,
	seekAudio,
	setAudioVolume,
	stopAudio
} from '#lib/player/audioEngine.ts';
import type { Crossfade } from '#lib/player/crossfade.svelte.ts';
import type { Player } from '#lib/player/player.svelte.ts';

/**
 * Connects the player to the Rust audio engine. The player's state goes out as commands (a new
 * `source` loads, `paused` plays or pauses, seeks jump, volume follows), and what the engine
 * reports comes back into the player. Call from a component's script: it registers effects.
 */
export function connectAudio(player: Player, crossfade: Crossfade): void {
	$effect(() => listen<AudioEvent>('audio', (event) => handleAudioEvent(player, crossfade, event)));
	$effect(() => loadWhenSourceChanges(player));
	$effect(() => playOrPauseWhenPausedChanges(player));
	$effect(() => player.onSeek((seconds) => seekAudio(mainSlot, seconds)));
	$effect(() => followVolume(player, crossfade));
	$effect(() => () => crossfade.finish());
}

function loadWhenSourceChanges(player: Player): void {
	const source = player.source;
	untrack(() => {
		if (source === undefined) {
			stopAudio(mainSlot);
			return;
		}
		loadAudio(mainSlot, source, player.currentTime, !player.paused);
	});
}

function playOrPauseWhenPausedChanges(player: Player): void {
	const paused = player.paused;
	untrack(() => {
		if (player.source === undefined) {
			return;
		}
		if (paused) {
			pauseAudio(mainSlot);
			return;
		}
		playAudio(mainSlot);
	});
}

// The fade-in scales the slot's volume without touching the user's volume setting.
function followVolume(player: Player, crossfade: Crossfade): void {
	let volume = player.volume * crossfade.incomingGain;
	if (player.muted) {
		volume = 0;
	}
	setAudioVolume(mainSlot, volume);
}

function handleAudioEvent(player: Player, crossfade: Crossfade, event: AudioEvent): void {
	if (event.slot === outgoingSlot) {
		crossfade.handleOutgoingEvent(event);
		return;
	}
	if (event.slot !== mainSlot) {
		return;
	}
	switch (event.kind) {
		case 'position':
			applyPosition(player, crossfade, event.positionSeconds);
			break;
		case 'duration':
			applyDuration(player, event.durationSeconds);
			break;
		case 'state':
			applyState(player, event.playing);
			break;
		case 'buffering':
			player.buffering = event.buffering;
			break;
		case 'ended':
			void player.trackEnded();
			break;
		case 'error':
			console.warn('audio error:', event.message);
			void player.playbackFailed();
			break;
	}
}

function applyPosition(player: Player, crossfade: Crossfade, positionSeconds: number): void {
	player.currentTime = positionSeconds;
	player.playbackAdvanced();
	crossfade.check();
}

// A song's own length is better than the one the queue knew, which can be 0 or rounded.
function applyDuration(player: Player, durationSeconds: number): void {
	player.duration = durationSeconds;
	if (player.paused) {
		player.buffering = false;
	}
}

function applyState(player: Player, playing: boolean): void {
	if (playing) {
		player.buffering = false;
	}
}
