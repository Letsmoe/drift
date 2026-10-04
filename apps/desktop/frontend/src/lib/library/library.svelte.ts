import { SvelteMap, SvelteSet } from 'svelte/reactivity';
import {
	addToPlaylist,
	createPlaylist,
	deletePlaylist,
	playlists,
	removeFromPlaylist,
	renamePlaylist,
	ratings,
	save,
	savedItems,
	setRating,
	unsave,
	type Playlist,
	type PlaylistSummary,
	type Rating,
	type SavedItem
} from '#lib/api/library.ts';
import type { Item, ItemKind, Track } from '#lib/api/music.ts';
import { describeError } from '#lib/format.ts';

/** The user's local library: ratings, playlists and saved items, kept in sync with the core. */
export class LibraryState {
	ratings = new SvelteMap<string, Rating>();
	playlists = $state<PlaylistSummary[]>([]);
	savedItems = $state<SavedItem[]>([]);
	errorMessage = $state('');

	#requestedRatingIds = new SvelteSet<string>();

	/** Loads playlists and saved items; call once at startup. */
	async load(): Promise<void> {
		await Promise.all([this.refreshPlaylists(), this.refreshSavedItems()]);
	}

	/** Loads the ratings of videoIds that have not been looked up yet. */
	async ensureRatings(videoIds: string[]): Promise<void> {
		const missingIds = videoIds.filter((videoId) => !this.#requestedRatingIds.has(videoId));
		if (missingIds.length === 0) {
			return;
		}
		for (const videoId of missingIds) {
			this.#requestedRatingIds.add(videoId);
		}
		try {
			const found = await ratings(missingIds);
			this.#storeRatings(missingIds, found);
		} catch (error) {
			this.#forgetRequested(missingIds);
			this.errorMessage = describeError(error);
		}
	}

	rating(videoId: string): Rating {
		const rating = this.ratings.get(videoId);
		if (rating === undefined) {
			return 0;
		}
		return rating;
	}

	isLiked(videoId: string): boolean {
		return this.rating(videoId) === 1;
	}

	isDisliked(videoId: string): boolean {
		return this.rating(videoId) === -1;
	}

	async toggleLike(track: Track, kind: ItemKind = 'song'): Promise<void> {
		if (this.isLiked(track.videoId)) {
			await this.#rate(track, kind, 0);
			return;
		}
		await this.#rate(track, kind, 1);
	}

	/** Dislikes the track, or clears the dislike when it already is disliked. */
	async dislike(track: Track, kind: ItemKind = 'song'): Promise<void> {
		if (this.isDisliked(track.videoId)) {
			await this.#rate(track, kind, 0);
			return;
		}
		await this.#rate(track, kind, -1);
	}

	async refreshPlaylists(): Promise<void> {
		try {
			this.playlists = await playlists();
		} catch (error) {
			this.errorMessage = describeError(error);
		}
	}

	async createPlaylist(name: string): Promise<Playlist> {
		const created = await createPlaylist(name);
		await this.refreshPlaylists();
		return created;
	}

	async addToPlaylist(id: string, tracks: Track[]): Promise<void> {
		await addToPlaylist(id, tracks);
		await this.refreshPlaylists();
	}

	async removeFromPlaylist(id: string, entryId: number): Promise<void> {
		await removeFromPlaylist(id, entryId);
		await this.refreshPlaylists();
	}

	/** Renames a playlist, keeping its description unless a new one is given. */
	async renamePlaylist(id: string, name: string, description?: string): Promise<void> {
		let newDescription = description;
		if (newDescription === undefined) {
			newDescription = this.#descriptionOf(id);
		}
		await renamePlaylist(id, name, newDescription);
		await this.refreshPlaylists();
	}

	async deletePlaylist(id: string): Promise<void> {
		await deletePlaylist(id);
		await this.refreshPlaylists();
	}

	async refreshSavedItems(): Promise<void> {
		try {
			this.savedItems = await savedItems();
		} catch (error) {
			this.errorMessage = describeError(error);
		}
	}

	isSaved(kind: ItemKind, id: string): boolean {
		return this.savedItems.some((saved) => saved.item.kind === kind && saved.item.id === id);
	}

	async toggleSaved(item: Item): Promise<void> {
		if (this.isSaved(item.kind as ItemKind, item.id)) {
			await unsave(item.kind as ItemKind, item.id);
		} else {
			await save(item);
		}
		await this.refreshSavedItems();
	}

	// Optimistic: the heart reacts at once and rolls back if the core rejects the rating.
	async #rate(track: Track, kind: ItemKind, rating: Rating): Promise<void> {
		const previousRating = this.rating(track.videoId);
		this.ratings.set(track.videoId, rating);
		try {
			await setRating(track, kind, rating);
		} catch (error) {
			this.ratings.set(track.videoId, previousRating);
			this.errorMessage = describeError(error);
		}
	}

	#storeRatings(requestedIds: string[], found: Record<string, Rating>): void {
		for (const videoId of requestedIds) {
			// A rating set while the lookup was in flight is newer than the lookup.
			if (!this.ratings.has(videoId)) {
				this.ratings.set(videoId, this.#foundRating(found, videoId));
			}
		}
	}

	#foundRating(found: Record<string, Rating>, videoId: string): Rating {
		const rating = found[videoId];
		if (rating === undefined) {
			return 0;
		}
		return rating;
	}

	#forgetRequested(videoIds: string[]): void {
		for (const videoId of videoIds) {
			this.#requestedRatingIds.delete(videoId);
		}
	}

	#descriptionOf(id: string): string {
		const summary = this.playlists.find((candidate) => candidate.id === id);
		if (!summary) {
			return '';
		}
		return summary.description;
	}
}

export const library = new LibraryState();
