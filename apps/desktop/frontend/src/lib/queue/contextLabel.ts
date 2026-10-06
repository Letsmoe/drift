const recommendContextPrefix = 'recommend:';
const storageKey = 'otter-music:context-titles';
// Enough for every list played recently; older names fall back to the generic label.
const maxStoredTitles = 200;

const fixedLabels: Record<string, string> = {
	home: 'Home',
	liked: 'Liked Songs',
	downloads: 'Downloads'
};

// Shown when a context's name was never seen, e.g. a queue restored from before this was added.
const kindLabels: Record<string, string> = {
	album: 'Album',
	playlist: 'Playlist',
	library: 'Playlist',
	artist: 'Artist',
	podcast: 'Podcast',
	song: 'Song radio'
};

const titles = loadTitles();

/** Remembers the name of a context (`album:<id>`, ...) so the queue can show it. */
export function rememberContextTitle(context: string, title: string): void {
	if (title === '' || titles.get(context) === title) {
		return;
	}
	titles.delete(context);
	titles.set(context, title);
	while (titles.size > maxStoredTitles) {
		titles.delete(titles.keys().next().value!);
	}
	localStorage.setItem(storageKey, JSON.stringify([...titles]));
}

/** Human readable name of the context a queue was started from. */
export function contextLabel(context: string): string {
	if (context === '') {
		return '';
	}
	if (context in fixedLabels) {
		return fixedLabels[context];
	}
	if (context.startsWith(recommendContextPrefix)) {
		return 'Recommended for you';
	}
	const title = titles.get(context);
	if (title !== undefined) {
		return title;
	}
	return kindLabel(context);
}

function kindLabel(context: string): string {
	const kind = context.split(':')[0];
	if (kind in kindLabels) {
		return kindLabels[kind];
	}
	return '';
}

function loadTitles(): Map<string, string> {
	try {
		const stored = localStorage.getItem(storageKey);
		if (stored === null) {
			return new Map();
		}
		return new Map(JSON.parse(stored) as [string, string][]);
	} catch {
		return new Map();
	}
}
