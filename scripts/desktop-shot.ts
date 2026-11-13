// Takes a screenshot of the desktop app on svelte-gpui's offscreen display, never on the real one.
//
//   bun scripts/desktop-shot.ts OUT.png SECONDS [--shell] [--track JSON]... [-- host flags]
//
// It builds the Go backend and the UI bundle, starts a backend of its own, and runs the plain
// svelte-gpui host against it, or with --shell the Rust shell (apps/desktop/shell, which starts its
// own backend and plays audio). Steps given after `--`, such as `--do press:Settings`, run 0.9 s apart,
// so SECONDS must cover them.
//
// Each --track is a music.Track as JSON, for example
//   --track '{"videoId":"dQw4w9WgXcQ","title":"Never Gonna Give You Up","artists":[{"id":"","name":"Rick Astley"}],"durationSeconds":213,"thumbnails":[]}'
// The tracks are put into the backend's queue about four seconds after the window opened, and the
// first one starts playing, as if it had been picked in the app.
//
// DRIFT_AUDIO_SINK=fakesink keeps the shell from making sound (it plays in real time to nowhere).
import { spawn, spawnSync, type ChildProcess } from 'node:child_process';
import { readFileSync, rmSync } from 'node:fs';
import { createInterface } from 'node:readline';
import path from 'node:path';

interface BackendAddress {
	url: string;
	token: string;
}

const repositoryDirectory = path.resolve(import.meta.dirname, '..');
const desktopDirectory = path.join(repositoryDirectory, 'apps/desktop');
const frontendDirectory = path.join(desktopDirectory, 'frontend');
const shellDirectory = path.join(desktopDirectory, 'shell');
const backendBinary = path.join(desktopDirectory, 'build/bin/drift-backend');
// Each run builds its own bundle, config directory and address file, so runs can go side by side.
const runId = String(process.pid);
const bundlePath = path.join(frontendDirectory, `dist/shot-${runId}.js`);
const addressFile = `/tmp/drift-shot-address-${runId}.json`;
const audioLogFile = `/tmp/drift-shot-audio-${runId}.log`;
const gpuiDirectory =
	process.env.SVELTE_GPUI_DIR || path.resolve(repositoryDirectory, '../svelte-gpui');
const announcePrefix = 'drift-backend ';
const queueDelayMs = 4000;

// The backend and the shell keep the library and the UI state in the config directory. A screenshot run
// gets a directory of its own, so it never reads or writes the library of an installed Drift.
process.env.XDG_CONFIG_HOME = process.env.DRIFT_SHOT_CONFIG || `/tmp/drift-shot-config-${runId}`;

interface Options {
	outputPath: string;
	seconds: string;
	useShell: boolean;
	tracks: unknown[];
	hostFlags: string[];
}

function parseArguments(): Options {
	const commandLine = process.argv.slice(2);
	const separatorIndex = commandLine.indexOf('--');
	const own = commandLine.slice(0, separatorIndex < 0 ? commandLine.length : separatorIndex);
	const hostFlags = separatorIndex < 0 ? [] : commandLine.slice(separatorIndex + 1);
	const positional: string[] = [];
	const tracks: unknown[] = [];
	for (let index = 0; index < own.length; index += 1) {
		if (own[index] === '--track') {
			index += 1;
			tracks.push(JSON.parse(own[index]));
			continue;
		}
		if (!own[index].startsWith('--')) {
			positional.push(own[index]);
		}
	}
	const [outputPath, seconds] = positional;
	if (!outputPath || !seconds) {
		console.error(
			'usage: bun scripts/desktop-shot.ts OUT.png SECONDS [--shell] [--track JSON]... [-- host flags]'
		);
		process.exit(2);
	}
	return { outputPath, seconds, useShell: own.includes('--shell'), tracks, hostFlags };
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

function startBackend(): { process: ChildProcess; address: Promise<BackendAddress> } {
	const backend = spawn(backendBinary, ['-exit-with-stdin'], {
		stdio: ['pipe', 'pipe', 'inherit'],
		env: process.env
	});
	const address = new Promise<BackendAddress>((resolve, reject) => {
		backend.once('exit', (code) => reject(new Error(`The backend exited (code ${code}).`)));
		createInterface({ input: backend.stdout! }).on('line', (line) => {
			if (line.startsWith(announcePrefix)) {
				resolve(JSON.parse(line.slice(announcePrefix.length)));
			}
		});
	});
	return { process: backend, address };
}

function runScreenshot(
	options: Options,
	host: string,
	environment: Record<string, string>
): Promise<number> {
	const child = spawn(
		path.join(gpuiDirectory, 'scripts/shot.sh'),
		[bundlePath, options.outputPath, options.seconds, '--dev', ...options.hostFlags],
		{ stdio: 'inherit', env: { ...process.env, ...environment, HOST: host } }
	);
	return new Promise((resolve) => child.on('exit', (code) => resolve(code || 0)));
}

// Puts the tracks into the queue over the backend's RPC socket and starts the first one.
function queueTracks(address: BackendAddress, tracks: unknown[]): Promise<void> {
	const socketUrl = `${address.url.replace('http', 'ws')}/rpc?token=${address.token}`;
	return new Promise((resolve, reject) => {
		const socket = new WebSocket(socketUrl);
		socket.onopen = () =>
			socket.send(
				JSON.stringify({ id: 1, method: 'QueuePlay', params: [tracks, 'song', 0, 'shot'] })
			);
		socket.onmessage = (message) => {
			const reply = JSON.parse(String(message.data));
			socket.close();
			if (reply.error) {
				reject(new Error(reply.error));
				return;
			}
			resolve();
		};
		socket.onerror = () => reject(new Error('could not reach the backend'));
	});
}

async function waitForAddressFile(): Promise<BackendAddress> {
	for (let attempt = 0; attempt < 100; attempt += 1) {
		try {
			return JSON.parse(readFileSync(addressFile, 'utf8'));
		} catch {
			await new Promise((resolve) => setTimeout(resolve, 200));
		}
	}
	throw new Error('the shell never published its backend address');
}

async function queueLater(address: Promise<BackendAddress>, tracks: unknown[]): Promise<void> {
	if (tracks.length === 0) {
		return;
	}
	await new Promise((resolve) => setTimeout(resolve, queueDelayMs));
	await queueTracks(await address, tracks);
}

// A failure to queue is reported, but the screenshot is still taken: it shows what went wrong.
function reportQueueFailure(error: unknown): void {
	console.error('could not queue the tracks:', error);
}

async function screenshotWithShell(options: Options): Promise<number> {
	runOrExit('bun', ['build.mjs'], frontendDirectory, { DRIFT_BUNDLE_OUT: bundlePath });
	runOrExit('cargo', ['build', '--release'], shellDirectory);
	rmSync(addressFile, { force: true });
	const shellBinary = path.join(shellDirectory, 'target/release/drift');
	const finished = runScreenshot(options, shellBinary, {
		DRIFT_BACKEND_BINARY: backendBinary,
		DRIFT_BACKEND_ADDRESS_FILE: addressFile,
		DRIFT_AUDIO_LOG: audioLogFile
	});
	await queueLater(waitForAddressFile(), options.tracks).catch(reportQueueFailure);
	const status = await finished;
	console.log(`audio events: ${audioLogFile}`);
	return status;
}

async function screenshotWithPlainHost(options: Options): Promise<number> {
	const backend = startBackend();
	const address = await backend.address;
	runOrExit('bun', ['build.mjs'], frontendDirectory, {
		DRIFT_BUNDLE_OUT: bundlePath,
		DRIFT_BACKEND_URL: address.url,
		DRIFT_BACKEND_TOKEN: address.token
	});
	const finished = runScreenshot(
		options,
		path.join(gpuiDirectory, 'target/release/svelte-gpui'),
		{}
	);
	await queueLater(Promise.resolve(address), options.tracks).catch(reportQueueFailure);
	const status = await finished;
	backend.process.kill();
	return status;
}

function cleanUp(): void {
	rmSync(bundlePath, { force: true });
	rmSync(`${bundlePath}.map`, { force: true });
	rmSync(addressFile, { force: true });
	// A DRIFT_SHOT_CONFIG directory is the caller's, such as one seeded with a ui-state.json.
	if (!process.env.DRIFT_SHOT_CONFIG) {
		rmSync(process.env.XDG_CONFIG_HOME!, { recursive: true, force: true });
	}
}

const options = parseArguments();
runOrExit('go', ['build', '-o', backendBinary, '.'], desktopDirectory);
let status: number;
if (options.useShell) {
	status = await screenshotWithShell(options);
} else {
	status = await screenshotWithPlainHost(options);
}
cleanUp();
process.exit(status);
