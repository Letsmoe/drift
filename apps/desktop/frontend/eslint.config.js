import neoworks from '@neoworks-dev/config/eslint';

export default [
	// build/ is the static output; src/lib/backend/generated is written by `bun run bindings`.
	{ ignores: ['build/', '.svelte-kit/', 'src/lib/backend/generated/'] },
	...neoworks
];
