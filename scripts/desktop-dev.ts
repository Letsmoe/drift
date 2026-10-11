// Runs the desktop app for development: the Vite dev server (hot reload for the
// UI), a fresh build of the Go backend and Electron pointed at both. Go changes
// need a restart of this script.
import { spawn, spawnSync, type ChildProcess } from 'node:child_process';
import { existsSync } from 'node:fs';
import net from 'node:net';
import { homedir } from 'node:os';
import path from 'node:path';

const desktopDirectory = path.resolve(import.meta.dirname, '../apps/desktop');
const frontendDirectory = path.join(desktopDirectory, 'frontend');
const electronDirectory = path.join(desktopDirectory, 'electron');
const backendBinary = path.join(desktopDirectory, 'build/bin/drift-backend');
const devServerPort = await findFreePort();
const devServerUrl = `http://localhost:${devServerPort}`;
const devServerTimeoutMs = 30_000;
// Development signs in to the local Neoworks stack (neoworks.dev's process-compose) unless told otherwise.
const localNeoworksDomain = 'neoworks.localhost:8443';
const localNeoworksAuthority = path.join(
	homedir(),
	'.local/share/caddy/pki/authorities/local/root.crt'
);

// Asks the OS for an unused port; --strictPort makes Vite fail loudly if it is taken in between.
function findFreePort(): Promise<number> {
	return new Promise((resolve, reject) => {
		const server = net.createServer();
		server.on('error', reject);
		server.listen(0, 'localhost', () => {
			const address = server.address() as net.AddressInfo;
			server.close(() => resolve(address.port));
		});
	});
}

function neoworksEnvironment(): Record<string, string> {
	const environment: Record<string, string> = { DRIFT_NEOWORKS_DOMAIN: localNeoworksDomain };
	if (existsSync(localNeoworksAuthority)) {
		environment.DRIFT_NEOWORKS_CA = localNeoworksAuthority;
	}
	return environment;
}

function runOrExit(command: string, args: string[], directory: string): void {
	const result = spawnSync(command, args, { cwd: directory, stdio: 'inherit' });
	if (result.status !== 0) {
		process.exit(result.status || 1);
	}
}

async function waitForDevServer(): Promise<void> {
	const deadline = Date.now() + devServerTimeoutMs;
	while (Date.now() < deadline) {
		try {
			await fetch(devServerUrl);
			return;
		} catch {
			await new Promise((resolve) => setTimeout(resolve, 250));
		}
	}
	throw new Error(`The Vite dev server did not answer on ${devServerUrl}.`);
}

function stopAll(children: ChildProcess[]): void {
	for (const child of children) {
		child.kill();
	}
}

runOrExit('go', ['build', '-o', backendBinary, '.'], desktopDirectory);
runOrExit('bun', ['run', 'build'], electronDirectory);

const vite = spawn('bun', ['run', 'dev', '--port', String(devServerPort), '--strictPort'], {
	cwd: frontendDirectory,
	stdio: 'inherit'
});
await waitForDevServer();

const electron = spawn('bun', ['x', 'electron', '.'], {
	cwd: electronDirectory,
	stdio: 'inherit',
	env: {
		...neoworksEnvironment(),
		...process.env,
		DRIFT_DEV_URL: devServerUrl,
		DRIFT_BACKEND_BINARY: backendBinary
	}
});

electron.on('exit', (code) => {
	stopAll([vite]);
	process.exit(code || 0);
});
for (const signal of ['SIGINT', 'SIGTERM'] as const) {
	process.on(signal, () => stopAll([electron, vite]));
}
