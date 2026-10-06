import {
	Album,
	Artist,
	Home,
	HomeContinuation,
	Playlist,
	PlaylistContinuation,
	Podcast,
	PodcastContinuation,
	PodcastShows,
	PodcastShowsContinuation,
	Radio,
	Search,
	SearchSongs,
	SearchSuggestions
} from '#lib/wailsjs/go/main/App.js';
import type {
	AlbumPage,
	ArtistPage,
	HomePage,
	PlaylistPage,
	PlaylistTracks,
	PodcastEpisodes,
	PodcastPage,
	SearchResults,
	Track
} from '#lib/api/music.ts';

export type {
	AlbumPage,
	ArtistPage,
	HomePage,
	PlaylistPage,
	PlaylistTracks,
	PodcastEpisodes,
	PodcastPage,
	SearchResults
};

export function searchSongs(query: string): Promise<Track[]> {
	return SearchSongs(query);
}

/** Searches everything and records the query in the search history. */
export function search(query: string): Promise<SearchResults> {
	return Search(query);
}

export function searchSuggestions(query: string): Promise<string[]> {
	return SearchSuggestions(query);
}

export function album(browseId: string): Promise<AlbumPage> {
	return Album(browseId);
}

export function playlist(browseId: string): Promise<PlaylistPage> {
	return Playlist(browseId);
}

export function playlistContinuation(token: string): Promise<PlaylistTracks> {
	return PlaylistContinuation(token);
}

export function artist(browseId: string): Promise<ArtistPage> {
	return Artist(browseId);
}

export function radio(videoId: string): Promise<Track[]> {
	return Radio(videoId);
}

export function podcast(browseId: string): Promise<PodcastPage> {
	return Podcast(browseId);
}

export function podcastContinuation(token: string): Promise<PodcastEpisodes> {
	return PodcastContinuation(token);
}

export function home(): Promise<HomePage> {
	return Home();
}

/** Podcast shows by topic (the shows behind YouTube's trending episodes). */
export function podcastShows(): Promise<HomePage> {
	return PodcastShows();
}

export function podcastShowsContinuation(continuation: string): Promise<HomePage> {
	return PodcastShowsContinuation(continuation);
}

export function homeContinuation(continuation: string): Promise<HomePage> {
	return HomeContinuation(continuation);
}
