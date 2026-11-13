/** What the app shell passes to every page: the `:name` segments of its path and its query string. */
export interface PageProps {
	params: Record<string, string>;
	query: Record<string, string>;
}
