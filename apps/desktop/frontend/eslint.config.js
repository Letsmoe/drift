import neoworks from '@neoworks-dev/config/eslint';

export default [
	// dist/ holds the bundles; src/lib/backend/generated is written by `bun run bindings`.
	{ ignores: ['dist/', 'src/lib/backend/generated/'] },
	...neoworks
];
