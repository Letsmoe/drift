// Bundles the Drift UI with svelte-gpui's build. The svelte-gpui checkout sits next to the drift
// checkout (neoworks/svelte-gpui); a worktree reaches it through .worktrees/svelte-gpui.
//
// DRIFT_BACKEND_URL and DRIFT_BACKEND_TOKEN are baked into the bundle, for running it with the
// plain `svelte-gpui` host against a backend that is already running. The Rust shell ignores
// them: it starts its own backend and answers `backend_address`.
import { readdir } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { build } from '../../../../svelte-gpui/build.mjs';

const here = (path) => fileURLToPath(new URL(path, import.meta.url));

const pagesDirectory = here('./src/pages');

// `virtual:pages` exports the pages that exist in src/pages as `{ Home: Component, ... }`, so the
// app builds while some pages are still unwritten and App.svelte shows a placeholder for those.
const pagesPlugin = {
	name: 'drift-pages',
	setup(build) {
		build.onResolve({ filter: /^virtual:pages$/ }, () => ({
			path: 'virtual:pages',
			namespace: 'drift-pages'
		}));
		build.onLoad({ filter: /.*/, namespace: 'drift-pages' }, async () => {
			const names = await listPageNames();
			const imports = names.map(
				(name) => `import ${name} from ${JSON.stringify(`${pagesDirectory}/${name}.svelte`)};`
			);
			const contents = `${imports.join('\n')}\nexport const pages = { ${names.join(', ')} };\n`;
			return { contents, loader: 'js', resolveDir: pagesDirectory, watchDirs: [pagesDirectory] };
		});
	}
};

async function listPageNames() {
	try {
		const entries = await readdir(pagesDirectory);
		return entries
			.filter((entry) => /^[A-Z][A-Za-z]*\.svelte$/.test(entry))
			.map((entry) => entry.replace(/\.svelte$/, ''));
	} catch {
		return [];
	}
}

await build({
	entry: here('./src/main.ts'),
	outfile: here('./dist/drift.js'),
	watch: process.argv.includes('--watch'),
	plugins: [pagesPlugin],
	define: {
		__DRIFT_BACKEND_URL__: JSON.stringify(process.env.DRIFT_BACKEND_URL || ''),
		__DRIFT_BACKEND_TOKEN__: JSON.stringify(process.env.DRIFT_BACKEND_TOKEN || '')
	}
});
