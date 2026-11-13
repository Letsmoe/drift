// The commands and events of the Rust shell (apps/desktop/shell), typed for invoke and listen.

/** What the audio engine reports. `slot` 0 is the playing song, 1 the song fading out. */
export type AudioEvent =
	| { kind: 'position'; slot: number; positionSeconds: number }
	| { kind: 'duration'; slot: number; durationSeconds: number }
	| { kind: 'state'; slot: number; playing: boolean }
	| { kind: 'buffering'; slot: number; buffering: boolean }
	| { kind: 'ended'; slot: number }
	| { kind: 'error'; slot: number; message: string };

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
		audio_set_volume: { args: { slot: number; volume: number }; result: null };
	}
	interface EventMap {
		audio: AudioEvent;
	}
}
