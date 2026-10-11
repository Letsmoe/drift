// Generates each @drift collection's descriptor into its generated/ directory.
// The descriptor is what `openschema publish` uploads and what core hands to
// libneoworks, so the two have to stay byte for byte the same.
import { spawnSync } from 'node:child_process';
import path from 'node:path';
import { collections, openschemaCli, schemaDirectory } from './neoworks-checkout.ts';

function generate(cli: string, collection: string): void {
	const directory = path.join(schemaDirectory, collection);
	const args = [cli, 'gen', `${collection}.schema`, '--target', 'descriptor', '--out', 'generated'];
	const result = spawnSync('bun', args, { cwd: directory, stdio: 'inherit' });
	if (result.status !== 0) {
		process.exit(result.status || 1);
	}
}

const cli = openschemaCli();
for (const collection of collections) {
	generate(cli, collection);
}
