// Runs the desktop app for development: the Vite dev server (hot reload for the
// UI), a fresh build of the Go backend and Electron pointed at both. Go changes
// need a restart of this script.
import { spawn, spawnSync, type ChildProcess } from 'node:child_process';
import path from 'node:path';

const desktopDirectory = path.resolve(import.meta.dirname, '../apps/desktop');
const frontendDirectory = path.join(desktopDirectory, 'frontend');
const electronDirectory = path.join(desktopDirectory, 'electron');
const backendBinary = path.join(desktopDirectory, 'build/bin/otter-music-backend');
const devServerUrl = 'http://localhost:5173';
const devServerTimeoutMs = 30_000;

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

const vite = spawn('bun', ['run', 'dev', '--strictPort'], {
	cwd: frontendDirectory,
	stdio: 'inherit'
});
await waitForDevServer();

const electron = spawn('bun', ['x', 'electron', '.'], {
	cwd: electronDirectory,
	stdio: 'inherit',
	env: { ...process.env, OTTER_DEV_URL: devServerUrl, OTTER_BACKEND_BINARY: backendBinary }
});

electron.on('exit', (code) => {
	stopAll([vite]);
	process.exit(code || 0);
});
for (const signal of ['SIGINT', 'SIGTERM'] as const) {
	process.on(signal, () => stopAll([electron, vite]));
}
