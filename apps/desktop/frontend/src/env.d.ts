declare module '*.svg' {
	const markup: string;
	export default markup;
}

/** Baked in by build.mjs from DRIFT_BACKEND_URL and DRIFT_BACKEND_TOKEN; empty when unset. */
declare const __DRIFT_BACKEND_URL__: string;
declare const __DRIFT_BACKEND_TOKEN__: string;

declare module 'virtual:pages' {
	import type { Component } from 'svelte';
	/** The components found in src/pages, by file name without the extension. */
	export const pages: Partial<
		Record<string, Component<{ params: Record<string, string>; query: Record<string, string> }>>
	>;
}
