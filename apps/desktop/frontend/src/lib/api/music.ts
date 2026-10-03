import { Home, HomeContinuation, SearchSongs, StreamURL } from '#lib/wailsjs/go/main/App.js';
import type { music } from '#lib/wailsjs/go/models.ts';

// Components import from here, never from the generated bindings directly.
export type Track = music.Track;
export type Artist = music.Artist;
export type Thumbnail = music.Thumbnail;
export type Item = music.Item;
export type Shelf = music.Shelf;
export type HomePage = music.HomePage;

export function searchSongs(query: string): Promise<Track[]> {
	return SearchSongs(query);
}

export function home(): Promise<HomePage> {
	return Home();
}

export function homeContinuation(continuation: string): Promise<HomePage> {
	return HomeContinuation(continuation);
}

/** URL of the core's local stream proxy for an <audio> element. */
export function streamUrl(videoId: string): Promise<string> {
	return StreamURL(videoId);
}
