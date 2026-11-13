// Takes a screenshot of the desktop app on svelte-gpui's offscreen display, never on the real one.
//
//   bun scripts/desktop-shot.ts OUT.png SECONDS [--shell] [-- host flags, such as --do press:Settings]
//
// It builds the Go backend and the UI bundle, starts a backend of its own, and runs the plain
// svelte-gpui host against it, or with --shell the Rust shell (apps/desktop/shell, which starts its
// own backend). Steps given with --do run 0.9 s apart, so SECONDS must cover them.
import { spawn, spawnSync, type ChildProcess } from 'node:child_process';
import { createInterface } from 'node:readline';
import path from 'node:path';

const repositoryDirectory = path.resolve(import.meta.dirname, '..');
const desktopDirectory = path.join(repositoryDirectory, 'apps/desktop');
const frontendDirectory = path.join(desktopDirectory, 'frontend');
const shellDirectory = path.join(desktopDirectory, 'shell');
const backendBinary = path.join(desktopDirectory, 'build/bin/drift-backend');
const bundlePath = path.join(frontendDirectory, 'dist/drift.js');
const gpuiDirectory =
	process.env.SVELTE_GPUI_DIR || path.resolve(repositoryDirectory, '../svelte-gpui');
const announcePrefix = 'drift-backend ';

const commandLine = process.argv.slice(2);
const separatorIndex = commandLine.indexOf('--');
const ownArguments = commandLine.slice(0, separatorIndex < 0 ? commandLine.length : separatorIndex);
const hostFlags = separatorIndex < 0 ? [] : commandLine.slice(separatorIndex + 1);
const useShell = ownArguments.includes('--shell');
const [outputPath, seconds] = ownArguments.filter((argument) => !argument.startsWith('--'));

if (!outputPath || !seconds) {
	console.error('usage: bun scripts/desktop-shot.ts OUT.png SECONDS [--shell] [-- host flags]');
	process.exit(2);
}

function runOrExit(command: string, args: string[], directory: string, environment = {}): void {
	const result = spawnSync(command, args, {
		cwd: directory,
		stdio: 'inherit',
		env: { ...process.env, ...environment }
	});
	if (result.status !== 0) {
		process.exit(result.status || 1);
	}
}

function startBackend(): {
	process: ChildProcess;
	address: Promise<{ url: string; token: string }>;
} {
	const backend = spawn(backendBinary, ['-exit-with-stdin'], {
		stdio: ['pipe', 'pipe', 'inherit'],
		env: process.env
	});
	const address = new Promise<{ url: string; token: string }>((resolve, reject) => {
		backend.once('exit', (code) => reject(new Error(`The backend exited (code ${code}).`)));
		createInterface({ input: backend.stdout! }).on('line', (line) => {
			if (line.startsWith(announcePrefix)) {
				resolve(JSON.parse(line.slice(announcePrefix.length)));
			}
		});
	});
	return { process: backend, address };
}

function screenshotWith(host: string, environment: Record<string, string>): number {
	const result = spawnSync(
		path.join(gpuiDirectory, 'scripts/shot.sh'),
		[bundlePath, outputPath, seconds, '--dev', ...hostFlags],
		{ stdio: 'inherit', env: { ...process.env, ...environment, HOST: host } }
	);
	return result.status || 0;
}

runOrExit('go', ['build', '-o', backendBinary, '.'], desktopDirectory);
if (useShell) {
	runOrExit('bun', ['build.mjs'], frontendDirectory);
	runOrExit('cargo', ['build', '--release'], shellDirectory);
	const shellBinary = path.join(shellDirectory, 'target/release/drift');
	process.exit(screenshotWith(shellBinary, { DRIFT_BACKEND_BINARY: backendBinary }));
}

const backend = startBackend();
const address = await backend.address;
runOrExit('bun', ['build.mjs'], frontendDirectory, {
	DRIFT_BACKEND_URL: address.url,
	DRIFT_BACKEND_TOKEN: address.token
});
const status = screenshotWith(path.join(gpuiDirectory, 'target/release/svelte-gpui'), {});
backend.process.kill();
process.exit(status);
