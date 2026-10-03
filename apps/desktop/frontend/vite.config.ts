import { realpathSync } from 'node:fs';
import path from 'node:path';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

// @neoworks-dev/ui is bun-linked from the neoworks.dev checkout, outside this repo.
const uiPackageDirectory = realpathSync(
	path.resolve(import.meta.dirname, 'node_modules/@neoworks-dev/ui')
);

function isDependency(filename: string): boolean {
	return filename.split(/[/\\]/).includes('node_modules');
}

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			// Wails serves the static build; every route falls back to the SPA shell.
			adapter: adapter({ fallback: 'index.html' }),
			compilerOptions: {
				runes: ({ filename }) => {
					if (isDependency(filename)) {
						return undefined;
					}
					return true;
				}
			}
		})
	],
	resolve: {
		// The linked ui package would otherwise pull svelte from its own node_modules.
		dedupe: ['svelte', 'phosphor-svelte']
	},
	server: {
		fs: {
			allow: [uiPackageDirectory]
		}
	}
});
