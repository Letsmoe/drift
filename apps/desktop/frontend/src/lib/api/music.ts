import { SearchSongs, StreamURL } from '#lib/wailsjs/go/main/App.js';
import type { music } from '#lib/wailsjs/go/models.ts';

// Components import from here, never from the generated bindings directly.
export type Track = music.Track;
export type Artist = music.Artist;
export type Thumbnail = music.Thumbnail;

export function searchSongs(query: string): Promise<Track[]> {
	return SearchSongs(query);
}

/** URL of the core's local stream proxy for an <audio> element. */
export function streamUrl(videoId: string): Promise<string> {
	return StreamURL(videoId);
}
