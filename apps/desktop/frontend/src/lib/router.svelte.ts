// In-memory routing with the paths the SvelteKit app used. The app shell (App.svelte) maps
// `router.page` to a component in src/pages; pages and components navigate with `goto`.

export type PageName =
	| 'Home'
	| 'Search'
	| 'Album'
	| 'Artist'
	| 'Playlist'
	| 'Podcast'
	| 'Song'
	| 'LibraryPlaylist'
	| 'CollectionTracks'
	| 'Downloads'
	| 'Settings'
	| 'NotFound';

interface RouteDefinition {
	page: PageName;
	/** Path segments; one that starts with a colon is a parameter. */
	segments: string[];
}

function defineRoute(page: PageName, pattern: string): RouteDefinition {
	return { page, segments: splitSegments(pattern) };
}

const routes: RouteDefinition[] = [
	defineRoute('Home', '/'),
	defineRoute('Search', '/search'),
	defineRoute('Album', '/album/:id'),
	defineRoute('Artist', '/artist/:id'),
	defineRoute('Playlist', '/playlist/:id'),
	defineRoute('Podcast', '/podcast/:id'),
	defineRoute('Song', '/song/:id'),
	defineRoute('LibraryPlaylist', '/library/playlist/:id'),
	defineRoute('CollectionTracks', '/collection/tracks'),
	defineRoute('Downloads', '/downloads'),
	defineRoute('Settings', '/settings')
];

interface Location {
	/** The path and query as given to `goto`, for the history. */
	href: string;
	path: string;
	page: PageName;
	params: Record<string, string>;
	query: Record<string, string>;
}

function splitSegments(path: string): string[] {
	return path.split('/').filter((segment) => segment !== '');
}

function safeDecode(text: string): string {
	try {
		return decodeURIComponent(text.replace(/\+/g, ' '));
	} catch {
		return text;
	}
}

function parseQuery(queryText: string): Record<string, string> {
	const query: Record<string, string> = {};
	for (const pair of queryText.split('&')) {
		if (pair === '') {
			continue;
		}
		const separatorIndex = pair.indexOf('=');
		if (separatorIndex < 0) {
			query[safeDecode(pair)] = '';
			continue;
		}
		query[safeDecode(pair.slice(0, separatorIndex))] = safeDecode(pair.slice(separatorIndex + 1));
	}
	return query;
}

function matchRoute(
	route: RouteDefinition,
	segments: string[]
): Record<string, string> | undefined {
	if (route.segments.length !== segments.length) {
		return undefined;
	}
	const params: Record<string, string> = {};
	for (let index = 0; index < segments.length; index += 1) {
		const expected = route.segments[index];
		if (expected.startsWith(':')) {
			params[expected.slice(1)] = safeDecode(segments[index]);
			continue;
		}
		if (expected !== segments[index]) {
			return undefined;
		}
	}
	return params;
}

function parseLocation(href: string): Location {
	const withoutHash = href.split('#')[0];
	const queryIndex = withoutHash.indexOf('?');
	let path = withoutHash;
	let queryText = '';
	if (queryIndex >= 0) {
		path = withoutHash.slice(0, queryIndex);
		queryText = withoutHash.slice(queryIndex + 1);
	}
	if (!path.startsWith('/')) {
		path = `/${path}`;
	}
	const segments = splitSegments(path);
	const query = parseQuery(queryText);
	for (const route of routes) {
		const params = matchRoute(route, segments);
		if (params) {
			return { href, path, page: route.page, params, query };
		}
	}
	return { href, path, page: 'NotFound', params: {}, query };
}

class Router {
	#history = $state<Location[]>([parseLocation('/')]);
	#position = $state(0);

	#current = $derived(this.#history[this.#position]);

	/** The path with its query; changes with every navigation, so it keys the page. */
	href = $derived(this.#current.href);
	/** The path without its query, such as `/album/MPREb_x`. */
	path = $derived(this.#current.path);
	page = $derived(this.#current.page);
	/** Values of the `:name` segments, decoded. */
	params = $derived(this.#current.params);
	query = $derived(this.#current.query);

	canGoBack = $derived(this.#position > 0);
	canGoForward = $derived(this.#position < this.#history.length - 1);

	goto(href: string): void {
		const next = parseLocation(href);
		if (next.href === this.#current.href) {
			return;
		}
		// A new entry drops everything that was ahead of it.
		this.#history = [...this.#history.slice(0, this.#position + 1), next];
		this.#position = this.#history.length - 1;
	}

	back(): void {
		if (this.canGoBack) {
			this.#position -= 1;
		}
	}

	forward(): void {
		if (this.canGoForward) {
			this.#position += 1;
		}
	}
}

export const router = new Router();

export function goto(path: string): void {
	router.goto(path);
}

export function back(): void {
	router.back();
}

export function forward(): void {
	router.forward();
}
