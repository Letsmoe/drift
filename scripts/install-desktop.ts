// Installs the built desktop app for the current user. Run `bun run desktop:build` first.
import { chmod, copyFile, cp, mkdir, rm, writeFile } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import { homedir } from 'node:os';
import path from 'node:path';

const appName = 'otter-music';
const buildDirectory = path.resolve(import.meta.dirname, '../apps/desktop/build/bin');

function requireBuild(buildPath: string): void {
	if (existsSync(buildPath)) {
		return;
	}
	console.error(`No build at ${buildPath}. Run \`bun run desktop:build\` first.`);
	process.exit(1);
}

function desktopEntry(binaryPath: string): string {
	return [
		'[Desktop Entry]',
		'Type=Application',
		'Name=Otter Music',
		'Comment=YouTube Music player',
		`Exec=${binaryPath}`,
		'Categories=AudioVideo;Audio;Player;',
		'Terminal=false',
		''
	].join('\n');
}

async function installLinux(): Promise<void> {
	const builtBinary = path.join(buildDirectory, appName);
	requireBuild(builtBinary);

	const binaryDirectory = path.join(homedir(), '.local/bin');
	const applicationsDirectory = path.join(homedir(), '.local/share/applications');
	const installedBinary = path.join(binaryDirectory, appName);

	await mkdir(binaryDirectory, { recursive: true });
	await mkdir(applicationsDirectory, { recursive: true });
	await copyFile(builtBinary, installedBinary);
	await chmod(installedBinary, 0o755);
	await writeFile(
		path.join(applicationsDirectory, `${appName}.desktop`),
		desktopEntry(installedBinary)
	);

	console.log(`Installed ${installedBinary} and its desktop entry.`);
}

async function installMacOS(): Promise<void> {
	const builtBundle = path.join(buildDirectory, `${appName}.app`);
	requireBuild(builtBundle);

	const installedBundle = path.join('/Applications', `${appName}.app`);
	await rm(installedBundle, { recursive: true, force: true });
	await cp(builtBundle, installedBundle, { recursive: true });

	console.log(`Installed ${installedBundle}.`);
}

if (process.platform === 'linux') {
	await installLinux();
} else if (process.platform === 'darwin') {
	await installMacOS();
} else {
	console.error(`Installing on ${process.platform} is not supported yet.`);
	process.exit(1);
}
