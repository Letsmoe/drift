// Installs the built desktop app for the current user. Run `bun run desktop:build` first.
import { copyFile, cp, mkdir, readdir, rm, symlink, writeFile } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import { homedir } from 'node:os';
import path from 'node:path';

const appName = 'otter-music';
const bundleName = 'Otter Music.app';
// electron-builder's unpacked output (`--dir`), see apps/desktop/electron/package.json.
const buildDirectory = path.resolve(import.meta.dirname, '../apps/desktop/build/electron');
const appIcon = path.resolve(import.meta.dirname, '../apps/desktop/build/appicon.png');

function requireBuild(buildPath: string): void {
	if (existsSync(buildPath)) {
		return;
	}
	console.error(`No build at ${buildPath}. Run \`bun run desktop:build\` first.`);
	process.exit(1);
}

function desktopEntry(binaryPath: string, iconPath: string): string {
	return [
		'[Desktop Entry]',
		'Type=Application',
		'Name=Otter Music',
		'Comment=YouTube Music player',
		`Exec=${binaryPath}`,
		`Icon=${iconPath}`,
		// Matches desktopName in apps/desktop/electron/package.json, so the window groups under this entry.
		'StartupWMClass=otter-music',
		'Categories=AudioVideo;Audio;Player;',
		'Terminal=false',
		''
	].join('\n');
}

async function installLinux(): Promise<void> {
	const builtApp = path.join(buildDirectory, 'linux-unpacked');
	requireBuild(builtApp);

	const binaryDirectory = path.join(homedir(), '.local/bin');
	const applicationsDirectory = path.join(homedir(), '.local/share/applications');
	const dataDirectory = path.join(homedir(), '.local/share', appName);
	const installedApp = path.join(dataDirectory, 'app');
	const installedBinary = path.join(binaryDirectory, appName);
	const installedIcon = path.join(dataDirectory, 'icon.png');

	await mkdir(binaryDirectory, { recursive: true });
	await mkdir(applicationsDirectory, { recursive: true });
	await rm(installedApp, { recursive: true, force: true });
	await mkdir(dataDirectory, { recursive: true });
	await cp(builtApp, installedApp, { recursive: true });
	await rm(installedBinary, { force: true });
	await symlink(path.join(installedApp, appName), installedBinary);
	await copyFile(appIcon, installedIcon);
	await writeFile(
		path.join(applicationsDirectory, `${appName}.desktop`),
		desktopEntry(installedBinary, installedIcon)
	);

	console.log(`Installed ${installedApp}, linked ${installedBinary} and wrote its desktop entry.`);
}

// electron-builder writes mac/ or mac-arm64/ depending on the architecture.
async function findMacBundle(): Promise<string> {
	const outputs = await readdir(buildDirectory).catch(() => []);
	const macOutput = outputs.find((name) => name.startsWith('mac'));
	if (!macOutput) {
		return path.join(buildDirectory, 'mac', bundleName);
	}
	return path.join(buildDirectory, macOutput, bundleName);
}

async function installMacOS(): Promise<void> {
	const builtBundle = await findMacBundle();
	requireBuild(builtBundle);

	const installedBundle = path.join('/Applications', bundleName);
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
