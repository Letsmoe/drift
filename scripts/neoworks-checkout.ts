// Where the @drift collections live, and the neoworks.dev checkout the Neoworks
// scripts take the openschema CLI and the local stack's ports from: NEOWORKS_DIR,
// or the directory next to this repository's main checkout.
import { spawnSync } from 'node:child_process';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';

export const schemaDirectory = path.resolve(import.meta.dirname, '../core/neoworks/schema');
export const collections = ['tracks', 'ratings', 'saved', 'playlists', 'history', 'blocks'];

function mainCheckoutDirectory(): string {
	const result = spawnSync('git', ['rev-parse', '--path-format=absolute', '--git-common-dir'], {
		cwd: import.meta.dirname,
		encoding: 'utf8'
	});
	if (result.status !== 0) {
		throw new Error(`git rev-parse failed: ${result.stderr}`);
	}
	return path.dirname(result.stdout.trim());
}

export function neoworksDirectory(): string {
	const configured = process.env.NEOWORKS_DIR;
	if (configured) {
		return path.resolve(configured);
	}
	return path.resolve(mainCheckoutDirectory(), '../neoworks.dev');
}

export function openschemaCli(): string {
	return path.join(neoworksDirectory(), 'packages/openschema/src/cli/index.ts');
}

/** The checkout's `.env`, which holds the stack slot's ports; the main checkout's has only some of them. */
function stackEnvironment(): Map<string, string> {
	const environment = new Map<string, string>();
	const file = path.join(neoworksDirectory(), '.env');
	if (!existsSync(file)) {
		return environment;
	}
	for (const line of readFileSync(file, 'utf8').split('\n')) {
		const match = /^([A-Z0-9_]+)=(.*)$/.exec(line.trim());
		if (match) {
			environment.set(match[1], match[2]);
		}
	}
	return environment;
}

/** A port of the local stack, from the checkout's `.env` or its default (slot 0). */
export function stackPort(name: string, defaultPort: number): number {
	const value = stackEnvironment().get(name);
	if (value === undefined) {
		return defaultPort;
	}
	return Number(value);
}
