// Generates each @drift collection's descriptor into its generated/ directory.
// The descriptor is what `openschema publish` uploads and what core hands to
// libneoworks, so the two have to stay byte for byte the same.
// The openschema CLI comes from a neoworks.dev checkout: NEOWORKS_DIR, or the
// directory next to this repository's main checkout.
import { spawnSync } from 'node:child_process';
import path from 'node:path';

const schemaDirectory = path.resolve(import.meta.dirname, '../core/neoworks/schema');
const collections = ['tracks', 'ratings', 'saved', 'playlists', 'history', 'blocks'];

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

function neoworksDirectory(): string {
	const configured = process.env.NEOWORKS_DIR;
	if (configured) {
		return path.resolve(configured);
	}
	return path.resolve(mainCheckoutDirectory(), '../neoworks.dev');
}

function generate(cli: string, collection: string): void {
	const directory = path.join(schemaDirectory, collection);
	const args = [cli, 'gen', `${collection}.schema`, '--target', 'descriptor', '--out', 'generated'];
	const result = spawnSync('bun', args, { cwd: directory, stdio: 'inherit' });
	if (result.status !== 0) {
		process.exit(result.status || 1);
	}
}

const cli = path.join(neoworksDirectory(), 'packages/openschema/src/cli/index.ts');
for (const collection of collections) {
	generate(cli, collection);
}
