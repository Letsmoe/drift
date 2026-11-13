// The commands and events of the Rust shell (apps/desktop/shell), typed for invoke and listen.

/** What the audio engine reports. `slot` 0 is the playing song, 1 the song fading out. */
export type AudioEvent =
	| { kind: 'position'; slot: number; positionSeconds: number }
	| { kind: 'duration'; slot: number; durationSeconds: number }
	| { kind: 'state'; slot: number; playing: boolean }
	| { kind: 'buffering'; slot: number; buffering: boolean }
	| { kind: 'ended'; slot: number }
	| { kind: 'error'; slot: number; message: string }
	/** The engine's crossfade is over: it ran out (`finished`) or a play, pause or seek ended it. */
	| { kind: 'crossfade'; how: 'finished' | 'cancelled' };

/** The song as MPRIS shows it (apps/desktop/shell/src/mpris.rs). */
export interface MprisTrack {
	/** Names the queue entry, so the same song twice in the queue is two tracks. */
	trackId: string;
	title: string;
	artists: string[];
	album: string;
	coverUrl: string;
	lengthSeconds: number;
}

/** Everything the UI tells MPRIS about its player. */
export interface MprisUpdate {
	track: MprisTrack | null;
	status: 'playing' | 'paused' | 'stopped';
	/** Where the player was when this was sent; MPRIS counts on from there while it plays. */
	positionSeconds: number;
	volume: number;
	canGoNext: boolean;
	canGoPrevious: boolean;
	/** The position changed by a seek, which MPRIS clients hear as `Seeked`. */
	seeked: boolean;
}

/** What media keys, widgets and `playerctl` ask the player to do. */
export type MprisRequest =
	| { kind: 'play' | 'pause' | 'playPause' | 'stop' | 'next' | 'previous' }
	| { kind: 'seek'; offsetSeconds: number }
	| { kind: 'setPosition'; positionSeconds: number }
	| { kind: 'setVolume'; volume: number };

declare module '@neoworks/svelte-gpui' {
	interface CommandMap {
		backend_address: { args: undefined; result: { url: string; token: string } };
		open_external: { args: { url: string }; result: null };
		audio_load: {
			args: { slot: number; url: string; positionSeconds: number; playing: boolean };
			result: null;
		};
		audio_play: { args: { slot: number }; result: null };
		audio_pause: { args: { slot: number }; result: null };
		audio_seek: { args: { slot: number; positionSeconds: number }; result: null };
		audio_stop: { args: { slot: number }; result: null };
		/** The user's volume (0 when muted) for every slot. */
		audio_set_volume: { args: { volume: number }; result: null };
		/** What a slot's volume is multiplied with; the crossfade moves it on its own. */
		audio_set_gain: { args: { slot: number; gain: number }; result: null };
		/** Fades slot 0 in and slot 1 out over `seconds`, then stops slot 1. */
		audio_crossfade: { args: { seconds: number }; result: null };
		audio_cancel_crossfade: { args: undefined; result: null };
		mpris_update: { args: MprisUpdate; result: null };
	}
	interface EventMap {
		audio: AudioEvent;
		mpris: MprisRequest;
	}
}
