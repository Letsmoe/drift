import { files } from '@neoworks/svelte-gpui';

// A stand-in for localStorage, which the host does not have. One JSON file in the config
// directory is read before the app mounts and kept in memory, so reads are synchronous;
// writes go back to the file in the background, one at a time.

const directoryName = 'drift';
const fileName = 'ui-state.json';

let values: Record<string, string> = {};
let filePath = '';
let writing = false;
let writeAgain = false;

/** Reads the state file. Call once, before anything that reads storage runs. */
export async function loadStorage(): Promise<void> {
	const { config } = await files.paths();
	const directory = `${config}/${directoryName}`;
	filePath = `${directory}/${fileName}`;
	await files.mkdir(directory, { recursive: true });
	values = await readValues(filePath);
}

async function readValues(path: string): Promise<Record<string, string>> {
	let text = '';
	try {
		text = await files.readText(path);
	} catch {
		// First start, or an unreadable file: begin empty.
		return {};
	}
	try {
		return JSON.parse(text) as Record<string, string>;
	} catch {
		return {};
	}
}

export function getItem(key: string): string | null {
	if (!(key in values)) {
		return null;
	}
	return values[key];
}

export function setItem(key: string, value: string): void {
	values[key] = value;
	scheduleWrite();
}

export function removeItem(key: string): void {
	delete values[key];
	scheduleWrite();
}

function scheduleWrite(): void {
	if (writing) {
		writeAgain = true;
		return;
	}
	void writeValues();
}

async function writeValues(): Promise<void> {
	writing = true;
	do {
		writeAgain = false;
		await writeFileOnce();
	} while (writeAgain);
	writing = false;
}

// Written beside the file and renamed into place, so a crash leaves the old file whole.
async function writeFileOnce(): Promise<void> {
	if (filePath === '') {
		return;
	}
	const temporaryPath = `${filePath}.tmp`;
	try {
		await files.writeText(temporaryPath, JSON.stringify(values));
		await files.rename(temporaryPath, filePath);
	} catch (error) {
		console.warn('could not save the UI state', error);
	}
}
