// Runs the desktop app for development: a fresh build of the Go backend, the UI bundle rebuilt
// on every save, and the window reloading on every build (svelte-gpui's --watch).
//
//   bun run desktop:dev            the Rust shell (apps/desktop/shell): starts the backend, plays audio
//   bun run desktop:dev --plain    the plain svelte-gpui host and a backend started here; no audio,
//                                  for working on pages and components
//
// Everything after `--` goes to the host, for example `-- --do press:Settings`. Go changes need a
// restart of this script.
import { spawn, spawnSync, type ChildProcess } from 'node:child_process';
import { existsSync } from 'node:fs';
import { createInterface } from 'node:readline';
import { homedir } from 'node:os';
import path from 'node:path';

const repositoryDirectory = path.resolve(import.meta.dirname, '..');
const desktopDirectory = path.join(repositoryDirectory, 'apps/desktop');
const frontendDirectory = path.join(desktopDirectory, 'frontend');
const shellDirectory = path.join(desktopDirectory, 'shell');
const backendBinary = path.join(desktopDirectory, 'build/bin/drift-backend');
const bundlePath = path.join(frontendDirectory, 'dist/drift.js');
// The svelte-gpui checkout sits next to this one (a worktree reaches it through .worktrees/).
const gpuiDirectory =
	process.env.SVELTE_GPUI_DIR || path.resolve(repositoryDirectory, '../svelte-gpui');
const plainHostBinary = path.join(gpuiDirectory, 'target/release/svelte-gpui');
const announcePrefix = 'drift-backend ';
// Development signs in to the local Neoworks stack (neoworks.dev's process-compose) unless told otherwise.
const localNeoworksDomain = 'neoworks.localhost:8443';
const localNeoworksAuthority = path.join(
	homedir(),
	'.local/share/caddy/pki/authorities/local/root.crt'
);

const commandLine = process.argv.slice(2);
const separatorIndex = commandLine.indexOf('--');
const ownFlags = commandLine.slice(0, separatorIndex < 0 ? commandLine.length : separatorIndex);
const hostFlags = separatorIndex < 0 ? [] : commandLine.slice(separatorIndex + 1);
const usePlainHost = ownFlags.includes('--plain');

const children: ChildProcess[] = [];

function neoworksEnvironment(): Record<string, string> {
	const environment: Record<string, string> = { DRIFT_NEOWORKS_DOMAIN: localNeoworksDomain };
	if (existsSync(localNeoworksAuthority)) {
		environment.DRIFT_NEOWORKS_CA = localNeoworksAuthority;
	}
	return environment;
}

function runOrExit(
	command: string,
	args: string[],
	directory: string,
	environment: Record<string, string> = {}
): void {
	const result = spawnSync(command, args, {
		cwd: directory,
		stdio: 'inherit',
		env: { ...process.env, ...environment }
	});
	if (result.status !== 0) {
		process.exit(result.status || 1);
	}
}

function startChild(
	command: string,
	args: string[],
	directory: string,
	environment = {}
): ChildProcess {
	const child = spawn(command, args, {
		cwd: directory,
		stdio: 'inherit',
		env: { ...neoworksEnvironment(), ...process.env, ...environment }
	});
	children.push(child);
	return child;
}

function stopAll(): void {
	for (const child of children) {
		child.kill();
	}
}

// Starts the backend and resolves with the address it announces on stdout.
function startBackend(): Promise<{ url: string; token: string }> {
	const backend = spawn(backendBinary, ['-exit-with-stdin'], {
		stdio: ['pipe', 'pipe', 'inherit'],
		env: { ...neoworksEnvironment(), ...process.env }
	});
	children.push(backend);
	return new Promise((resolve, reject) => {
		backend.once('exit', (code) => reject(new Error(`The backend exited (code ${code}).`)));
		createInterface({ input: backend.stdout! }).on('line', (line) => {
			if (line.startsWith(announcePrefix)) {
				resolve(JSON.parse(line.slice(announcePrefix.length)));
				return;
			}
			console.log(line);
		});
	});
}

async function runPlain(): Promise<ChildProcess> {
	if (!existsSync(plainHostBinary)) {
		throw new Error(
			`No svelte-gpui host at ${plainHostBinary}; build it with scripts/host.sh there.`
		);
	}
	const address = await startBackend();
	const addressEnvironment = { DRIFT_BACKEND_URL: address.url, DRIFT_BACKEND_TOKEN: address.token };
	// The host needs a bundle to start from; the watching build then keeps it current.
	runOrExit('bun', ['build.mjs'], frontendDirectory, addressEnvironment);
	startChild('bun', ['build.mjs', '--watch'], frontendDirectory, addressEnvironment);
	return startChild(
		plainHostBinary,
		['--bundle', bundlePath, '--watch', ...hostFlags],
		frontendDirectory
	);
}

function runShell(): ChildProcess {
	runOrExit('bun', ['build.mjs'], frontendDirectory);
	startChild('bun', ['build.mjs', '--watch'], frontendDirectory);
	return startChild(
		'cargo',
		['run', '--release', '--', '--bundle', bundlePath, '--watch', ...hostFlags],
		shellDirectory,
		{ DRIFT_BACKEND_BINARY: backendBinary }
	);
}

runOrExit('go', ['build', '-o', backendBinary, '.'], desktopDirectory);
let host: ChildProcess;
if (usePlainHost) {
	host = await runPlain();
} else {
	host = runShell();
}

host.on('exit', (code) => {
	stopAll();
	process.exit(code || 0);
});
for (const signal of ['SIGINT', 'SIGTERM'] as const) {
	process.on(signal, () => {
		stopAll();
	});
}
