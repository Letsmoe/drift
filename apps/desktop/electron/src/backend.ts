import { spawn, type ChildProcess } from 'node:child_process';
import { createInterface } from 'node:readline';

/** Where the backend listens and the token its RPC socket requires; see apps/desktop/main.go. */
export interface BackendAddress {
	url: string;
	token: string;
}

export interface Backend {
	address: BackendAddress;
	process: ChildProcess;
	/** Closes the backend's stdin (it then shuts down cleanly) and kills it if it lingers. */
	stop(): void;
}

const announcePrefix = 'otter-backend ';
const startTimeoutMs = 30_000;
const stopTimeoutMs = 3_000;

export function startBackend(binaryPath: string): Promise<Backend> {
	const child = spawn(binaryPath, ['-exit-with-stdin'], { stdio: ['pipe', 'pipe', 'inherit'] });
	return new Promise((resolve, reject) => {
		const timeout = setTimeout(() => {
			child.kill();
			reject(new Error(`The backend did not start within ${startTimeoutMs / 1000} s.`));
		}, startTimeoutMs);

		child.once('error', (error) => {
			clearTimeout(timeout);
			reject(new Error(`Could not start the backend at ${binaryPath}: ${error.message}`));
		});
		child.once('exit', (code) => {
			clearTimeout(timeout);
			reject(new Error(`The backend exited during startup (code ${code}).`));
		});

		readAnnouncement(child, (address) => {
			clearTimeout(timeout);
			resolve({ address, process: child, stop: () => stopBackend(child) });
		});
	});
}

// Calls onAnnounce once with the address; other stdout lines pass through.
function readAnnouncement(
	child: ChildProcess,
	onAnnounce: (address: BackendAddress) => void
): void {
	if (!child.stdout) {
		return;
	}
	const lines = createInterface({ input: child.stdout });
	lines.on('line', (line) => {
		if (!line.startsWith(announcePrefix)) {
			console.log(line);
			return;
		}
		onAnnounce(JSON.parse(line.slice(announcePrefix.length)) as BackendAddress);
	});
}

function stopBackend(child: ChildProcess): void {
	if (child.exitCode !== null) {
		return;
	}
	child.stdin?.end();
	const killTimer = setTimeout(() => child.kill(), stopTimeoutMs);
	child.once('exit', () => clearTimeout(killTimer));
}
