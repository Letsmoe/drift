import { artist as fetchArtist } from '#lib/api/catalog.ts';
import type { ArtistPage } from '#lib/api/music.ts';

/** Artist pages by browse ID, fetched once per session; undefined until loaded or when it failed. */
class ArtistInfo {
	#pages = $state<Record<string, ArtistPage>>({});
	#requested = new Set<string>();

	pageOf(browseId: string): ArtistPage | undefined {
		return this.#pages[browseId];
	}

	async load(browseId: string): Promise<void> {
		if (browseId === '' || this.#requested.has(browseId)) {
			return;
		}
		this.#requested.add(browseId);
		try {
			const page = await fetchArtist(browseId);
			this.#pages = { ...this.#pages, [browseId]: page };
		} catch {
			// Offline or unknown artist: the panel just leaves the card out. Retry next time.
			this.#requested.delete(browseId);
		}
	}
}

export const artistInfo = new ArtistInfo();
