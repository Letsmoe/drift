import { StreamURL } from '#lib/wailsjs/go/main/App.js';
import type { music } from '#lib/wailsjs/go/models.ts';

// Components import from here, never from the generated bindings directly.
export type Track = music.Track;
export type Artist = music.Artist;
export type Album = music.Album;
export type Thumbnail = music.Thumbnail;
export type Item = music.Item;
export type Shelf = music.Shelf;
export type HomePage = music.HomePage;
export type SearchResults = music.SearchResults;
export type AlbumPage = music.AlbumPage;
export type PlaylistPage = music.PlaylistPage;
export type PlaylistTracks = music.PlaylistTracks;
export type ArtistPage = music.ArtistPage;
export type PodcastPage = music.PodcastPage;
export type PodcastEpisodes = music.PodcastEpisodes;
export type Episode = music.Episode;

/** What an item is: songs and episodes play, the others open a page. */
export type ItemKind = 'song' | 'album' | 'playlist' | 'artist' | 'podcast' | 'episode';

/** URL of the core's local stream proxy for an <audio> element. */
export function streamUrl(videoId: string): Promise<string> {
	return StreamURL(videoId);
}
