import { invoke } from '@neoworks/svelte-gpui';
import type { CommandMap } from '@neoworks/svelte-gpui';

// Thin calls into the Rust audio engine (apps/desktop/shell/src/audio.rs). Slot 0 plays the song;
// slot 1 plays the ending song while a crossfade runs.

export const mainSlot = 0;
export const outgoingSlot = 1;

type AudioCommand = Extract<keyof CommandMap, `audio_${string}`>;

// A failing engine (no GStreamer, a bad stream) shows as an error event or a log line, never a throw
// into the player's effects.
function send<Name extends AudioCommand>(name: Name, args: CommandMap[Name]['args']): void {
	invoke(name, args).catch((error: unknown) => console.warn(`${name} failed`, error));
}

export function loadAudio(
	slot: number,
	url: string,
	positionSeconds: number,
	playing: boolean
): void {
	send('audio_load', { slot, url, positionSeconds, playing });
}

export function playAudio(slot: number): void {
	send('audio_play', { slot });
}

export function pauseAudio(slot: number): void {
	send('audio_pause', { slot });
}

export function seekAudio(slot: number, positionSeconds: number): void {
	send('audio_seek', { slot, positionSeconds });
}

export function stopAudio(slot: number): void {
	send('audio_stop', { slot });
}

export function setAudioVolume(slot: number, volume: number): void {
	send('audio_set_volume', { slot, volume });
}
