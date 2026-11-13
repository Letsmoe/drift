// Builds the desktop app and installs it for the current user (Linux).
//
//   bun run desktop:install
//
// The Go backend is built, then svelte-gpui's packager turns the shell (apps/desktop/shell, an app
// with its own Rust crate) and the UI bundle into one executable. Both go to ~/.local/share/drift,
// the shell finds the backend next to itself, and ~/.local/bin/drift links to the shell. The desktop
// entry is called `drift`, which is the name the shell's MPRIS player gives as its DesktopEntry.
//
// HOME decides where it installs, so a trial run can use a temporary one (cargo and rustup then need
// their real directories: CARGO_HOME=$HOME/.cargo RUSTUP_HOME=$HOME/.rustup).
import { spawnSync } from 'node:child_process';
import { copyFile, mkdir, rm, symlink, writeFile } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import { homedir } from 'node:os';
import path from 'node:path';

const applicationName = 'drift';
// The window's application id, set by the packager from the "svelte-gpui" key of the frontend's package.json.
const windowApplicationId = 'dev.letsmoe.drift';

const repositoryDirectory = path.resolve(import.meta.dirname, '..');
const desktopDirectory = path.join(repositoryDirectory, 'apps/desktop');
const frontendDirectory = path.join(desktopDirectory, 'frontend');
const shellTargetDirectory = path.join(desktopDirectory, 'shell/target');
const packageDirectory = path.join(desktopDirectory, 'build/package');
const backendBinary = path.join(desktopDirectory, 'build/bin/drift-backend');
const appIcon = path.join(desktopDirectory, 'build/appicon.png');
const gpuiDirectory =
	process.env.SVELTE_GPUI_DIR || path.resolve(repositoryDirectory, '../svelte-gpui');

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

function buildBackend(): void {
	runOrExit(
		'go',
		['build', '-trimpath', '-ldflags=-s -w', '-o', backendBinary, '.'],
		desktopDirectory
	);
}

// The packager runs the frontend's build, builds the shell with its `package` feature and collects
// the executable in packageDirectory/frontend. It builds into the shell's target directory, which
// already holds the compiled GPUI.
function packageShell(): string {
	const packager = path.join(gpuiDirectory, 'scripts/package.mjs');
	if (!existsSync(packager)) {
		console.error(`No svelte-gpui packager at ${packager}; check out svelte-gpui next to drift.`);
		process.exit(1);
	}
	runOrExit(
		'bun',
		[packager, '--app', frontendDirectory, '--out', packageDirectory, '--no-tar'],
		repositoryDirectory,
		{ CARGO_TARGET_DIR: shellTargetDirectory }
	);
	return path.join(packageDirectory, 'frontend', applicationName);
}

function desktopEntry(executablePath: string): string {
	return [
		'[Desktop Entry]',
		'Type=Application',
		'Name=Drift',
		'Comment=YouTube Music player',
		`Exec=${executablePath}`,
		`Icon=${applicationName}`,
		`StartupWMClass=${windowApplicationId}`,
		'Categories=AudioVideo;Audio;Player;',
		'Terminal=false',
		''
	].join('\n');
}

async function install(shellBinary: string): Promise<void> {
	const home = homedir();
	const binaryDirectory = path.join(home, '.local/bin');
	const applicationsDirectory = path.join(home, '.local/share/applications');
	const iconDirectory = path.join(home, '.local/share/icons/hicolor/1024x1024/apps');
	const dataDirectory = path.join(home, '.local/share', applicationName);
	const installedShell = path.join(dataDirectory, applicationName);
	const linkedShell = path.join(binaryDirectory, applicationName);

	await mkdir(binaryDirectory, { recursive: true });
	await mkdir(applicationsDirectory, { recursive: true });
	await mkdir(iconDirectory, { recursive: true });
	await mkdir(dataDirectory, { recursive: true });
	// The Electron app lived in app/ and replaced the link; the shell has no use for it.
	await rm(path.join(dataDirectory, 'app'), { recursive: true, force: true });
	await rm(installedShell, { force: true });
	await copyFile(shellBinary, installedShell);
	await rm(path.join(dataDirectory, 'drift-backend'), { force: true });
	await copyFile(backendBinary, path.join(dataDirectory, 'drift-backend'));
	await rm(linkedShell, { force: true });
	await symlink(installedShell, linkedShell);
	await copyFile(appIcon, path.join(iconDirectory, `${applicationName}.png`));
	await writeFile(
		path.join(applicationsDirectory, `${applicationName}.desktop`),
		desktopEntry(linkedShell)
	);
	console.log(
		`Installed ${installedShell} and drift-backend, linked ${linkedShell} and wrote ${applicationName}.desktop.`
	);
}

if (process.platform !== 'linux') {
	console.error(`Installing on ${process.platform} is not supported yet.`);
	process.exit(1);
}
buildBackend();
await install(packageShell());
