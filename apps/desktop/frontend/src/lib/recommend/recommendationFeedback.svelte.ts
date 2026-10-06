import { SvelteSet } from 'svelte/reactivity';
import type { Artist, ItemKind, Track } from '#lib/api/music.ts';
import { blockRecommendation } from '#lib/api/recommend.ts';
import { library } from '#lib/library/library.svelte.ts';
import { player } from '#lib/player/player.svelte.ts';
import { splitQueue } from '#lib/queue/queueSections.ts';

/**
 * Negative feedback on recommendations. Core keeps the choice; the sets here hide
 * the tracks from lists already on screen until they are reloaded.
 */
class RecommendationFeedback {
	#hiddenTracks = new SvelteSet<string>();
	#hiddenArtists = new SvelteSet<string>();
	#hiddenShows = new SvelteSet<string>();

	/** Whether a recommended track should no longer be shown. */
	hides(track: Track): boolean {
		if (this.#hiddenTracks.has(track.videoId) || library.isDisliked(track.videoId)) {
			return true;
		}
		if (track.album && this.#hiddenShows.has(track.album.id)) {
			return true;
		}
		return track.artists.some((artist) => artist.id !== '' && this.#hiddenArtists.has(artist.id));
	}

	/** Dislikes the track (a negative taste signal), or lifts the dislike. */
	async toggleDislike(track: Track, kind: ItemKind): Promise<void> {
		const wasDisliked = library.isDisliked(track.videoId);
		await library.dislike(track, kind);
		if (!wasDisliked) {
			await this.#dropMatchingAutoplay();
		}
	}

	/** Never recommends the track again, without counting it against its artist. */
	async blockTrack(track: Track): Promise<void> {
		await blockRecommendation('track', track.videoId, track.title);
		this.#hiddenTracks.add(track.videoId);
		await this.#dropMatchingAutoplay();
	}

	/** Never recommends anything by the artist again. */
	async blockArtist(artist: Artist): Promise<void> {
		await blockRecommendation('artist', artist.id, artist.name);
		this.#hiddenArtists.add(artist.id);
		await this.#dropMatchingAutoplay();
	}

	// Only autoplay entries go: songs the user queued on purpose stay.
	/** Never recommends episodes of the podcast show again. */
	async blockPodcast(showId: string, title: string): Promise<void> {
		await blockRecommendation('podcast', showId, title);
		this.#hiddenShows.add(showId);
		await this.#dropMatchingAutoplay();
	}

	async #dropMatchingAutoplay(): Promise<void> {
		const { autoplay } = splitQueue(player.queue);
		for (const upcoming of autoplay) {
			if (this.hides(upcoming.entry.track)) {
				await player.remove(upcoming.entry.id);
			}
		}
	}
}

export const recommendationFeedback = new RecommendationFeedback();
