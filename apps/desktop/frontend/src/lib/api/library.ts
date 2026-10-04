import {
	AddToPlaylist,
	ClearSearches,
	CreatePlaylist,
	DeletePlaylist,
	IsSaved,
	LibraryPlaylist,
	LikedTracks,
	MovePlaylistEntry,
	Playlists,
	Ratings,
	RecentSearches,
	RecentTracks,
	RecordImpressions,
	RemoveFromPlaylist,
	RenamePlaylist,
	ResumePosition,
	Save,
	SavedItems,
	SetRating,
	StartPlay,
	Unsave,
	UpdatePlay
} from '#lib/wailsjs/go/main/App.js';
import type { library } from '#lib/wailsjs/go/models.ts';
import type { Item, ItemKind, Track } from '#lib/api/music.ts';

export type Playlist = library.Playlist;
export type PlaylistEntry = library.PlaylistEntry;
export type PlaylistSummary = library.PlaylistSummary;
export type SavedItem = library.SavedItem;
export type Impression = library.Impression;

/** A track rating: 1 like, -1 dislike, 0 none. */
export type Rating = -1 | 0 | 1;

export interface ResumePoint {
	seconds: number;
	found: boolean;
}

export function setRating(track: Track, kind: ItemKind, rating: Rating): Promise<void> {
	return SetRating(track, kind, rating);
}

/** Ratings of the given videos as a map; unrated videos are absent. */
export async function ratings(videoIds: string[]): Promise<Record<string, Rating>> {
	const found = await Ratings(videoIds);
	return found as Record<string, Rating>;
}

export function likedTracks(): Promise<Track[]> {
	return LikedTracks();
}

export function recentTracks(limit: number): Promise<Track[]> {
	return RecentTracks(limit);
}

export function playlists(): Promise<PlaylistSummary[]> {
	return Playlists();
}

export function libraryPlaylist(id: string): Promise<Playlist> {
	return LibraryPlaylist(id);
}

export function createPlaylist(name: string): Promise<Playlist> {
	return CreatePlaylist(name);
}

export function renamePlaylist(id: string, name: string, description: string): Promise<void> {
	return RenamePlaylist(id, name, description);
}

export function deletePlaylist(id: string): Promise<void> {
	return DeletePlaylist(id);
}

export function addToPlaylist(id: string, tracks: Track[]): Promise<void> {
	return AddToPlaylist(id, tracks);
}

export function removeFromPlaylist(id: string, entryId: number): Promise<void> {
	return RemoveFromPlaylist(id, entryId);
}

export function movePlaylistEntry(id: string, entryId: number, newIndex: number): Promise<void> {
	return MovePlaylistEntry(id, entryId, newIndex);
}

export function save(item: Item): Promise<void> {
	return Save(item);
}

export function unsave(kind: ItemKind, id: string): Promise<void> {
	return Unsave(kind, id);
}

export function isSaved(kind: ItemKind, id: string): Promise<boolean> {
	return IsSaved(kind, id);
}

export function savedItems(): Promise<SavedItem[]> {
	return SavedItems();
}

/** Records the start of a playback; the returned play ID goes to updatePlay. */
export function startPlay(track: Track, kind: ItemKind, playContext: string): Promise<number> {
	return StartPlay(track, kind, playContext);
}

export function updatePlay(
	playId: number,
	progressSeconds: number,
	completed: boolean
): Promise<void> {
	return UpdatePlay(playId, progressSeconds, completed);
}

/** Where an unfinished episode picks up; `found` is false when there is nothing to resume. */
export function resumePosition(videoId: string): Promise<ResumePoint> {
	return ResumePosition(videoId);
}

export function recentSearches(limit: number): Promise<string[]> {
	return RecentSearches(limit);
}

export function clearSearches(): Promise<void> {
	return ClearSearches();
}

export function recordImpressions(impressions: Impression[]): Promise<void> {
	return RecordImpressions(impressions);
}
