import path from 'node:path';
import {
	app,
	BrowserWindow,
	dialog,
	ipcMain,
	Menu,
	net,
	protocol,
	shell,
	type OpenDialogOptions
} from 'electron';
import { startBackend, type Backend } from './backend.ts';

// The page lives on otter://app so its origin (and localStorage) stays the same
// although the backend picks a new port every run; requests are forwarded to it.
const appScheme = 'otter';
const appOrigin = `${appScheme}://app`;

// Set by scripts/desktop-dev.ts: the Vite dev server and a freshly built backend.
const devServerUrl = process.env.OTTER_DEV_URL;
const devBackendBinary = process.env.OTTER_BACKEND_BINARY;

const backendBinaryName =
	process.platform === 'win32' ? 'otter-music-backend.exe' : 'otter-music-backend';

let backend: Backend | undefined;
let mainWindow: BrowserWindow | undefined;
let quitting = false;

protocol.registerSchemesAsPrivileged([
	{ scheme: appScheme, privileges: { standard: true, secure: true, supportFetchAPI: true } }
]);
// Playback continues from the queue without a click on the page.
app.commandLine.appendSwitch('autoplay-policy', 'no-user-gesture-required');

function resourcePath(...parts: string[]): string {
	if (app.isPackaged) {
		return path.join(process.resourcesPath, ...parts);
	}
	// Unpackaged, the app path is apps/desktop/electron and the backend builds into ../build/bin.
	return path.join(app.getAppPath(), '..', 'build', ...parts);
}

function backendBinaryPath(): string {
	if (devBackendBinary) {
		return devBackendBinary;
	}
	if (app.isPackaged) {
		return resourcePath('backend', backendBinaryName);
	}
	return resourcePath('bin', backendBinaryName);
}

function iconPath(): string {
	if (app.isPackaged) {
		return resourcePath('icon.png');
	}
	return resourcePath('appicon.png');
}

function requireBackend(): Backend {
	if (!backend) {
		throw new Error('The backend is not running.');
	}
	return backend;
}

// Forwards otter://app/<path> to the backend, which serves the SPA build.
function serveAppScheme(): void {
	protocol.handle(appScheme, (request) => {
		const requested = new URL(request.url);
		const target = new URL(requested.pathname + requested.search, requireBackend().address.url);
		return net.fetch(target.toString());
	});
}

function registerIpc(): void {
	ipcMain.on('otter:backend', (event) => {
		event.returnValue = requireBackend().address;
	});
	ipcMain.handle('otter:pick-file', (_event, options: OpenDialogOptions) =>
		pickPath({ ...options, properties: ['openFile'] })
	);
	ipcMain.handle('otter:pick-folder', (_event, options: OpenDialogOptions) =>
		pickPath({ ...options, properties: ['openDirectory'] })
	);
}

async function pickPath(options: OpenDialogOptions): Promise<string> {
	if (!mainWindow) {
		return '';
	}
	const result = await dialog.showOpenDialog(mainWindow, options);
	if (result.canceled || result.filePaths.length === 0) {
		return '';
	}
	return result.filePaths[0];
}

function isAppUrl(url: string): boolean {
	if (devServerUrl) {
		return url.startsWith(devServerUrl);
	}
	return url.startsWith(appOrigin);
}

// Links leaving the app open in the default browser instead of the window.
function keepNavigationInApp(window: BrowserWindow): void {
	window.webContents.setWindowOpenHandler(({ url }) => {
		openExternally(url);
		return { action: 'deny' };
	});
	window.webContents.on('will-navigate', (event, url) => {
		if (isAppUrl(url)) {
			return;
		}
		event.preventDefault();
		openExternally(url);
	});
}

function openExternally(url: string): void {
	if (url.startsWith('https://') || url.startsWith('http://')) {
		void shell.openExternal(url);
	}
}

function createWindow(): void {
	const window = new BrowserWindow({
		title: 'Otter Music',
		width: 1200,
		height: 800,
		minWidth: 720,
		minHeight: 480,
		backgroundColor: '#141416',
		icon: iconPath(),
		show: false,
		autoHideMenuBar: true,
		webPreferences: {
			preload: path.join(import.meta.dirname, 'preload.cjs'),
			contextIsolation: true,
			sandbox: true
		}
	});
	window.once('ready-to-show', () => window.show());
	window.on('closed', () => {
		mainWindow = undefined;
	});
	keepNavigationInApp(window);
	toggleDevToolsOnShortcut(window);
	void window.loadURL(devServerUrl || `${appOrigin}/`);
	mainWindow = window;
}

// The removed application menu used to provide these shortcuts.
function toggleDevToolsOnShortcut(window: BrowserWindow): void {
	window.webContents.on('before-input-event', (event, input) => {
		if (input.type !== 'keyDown' || !isDevToolsShortcut(input)) {
			return;
		}
		event.preventDefault();
		window.webContents.toggleDevTools();
	});
}

function isDevToolsShortcut(input: Electron.Input): boolean {
	if (input.key === 'F12') {
		return true;
	}
	const isKeyI = input.key.toLowerCase() === 'i';
	if (process.platform === 'darwin') {
		return isKeyI && input.meta && input.alt;
	}
	return isKeyI && input.control && input.shift;
}

function quitWhenBackendDies(running: Backend): void {
	running.process.once('exit', (code) => {
		if (quitting || code === 0) {
			app.quit();
			return;
		}
		dialog.showErrorBox('Otter Music', `The backend stopped unexpectedly (code ${code}).`);
		app.quit();
	});
}

async function start(): Promise<void> {
	try {
		backend = await startBackend(backendBinaryPath());
	} catch (error) {
		dialog.showErrorBox('Otter Music', String(error));
		app.quit();
		return;
	}
	quitWhenBackendDies(backend);
	serveAppScheme();
	registerIpc();
	createWindow();
}

function focusMainWindow(): void {
	if (!mainWindow) {
		return;
	}
	if (mainWindow.isMinimized()) {
		mainWindow.restore();
	}
	mainWindow.focus();
}

if (!app.requestSingleInstanceLock()) {
	app.quit();
} else {
	app.on('second-instance', focusMainWindow);
	app.on('window-all-closed', () => app.quit());
	app.on('before-quit', () => {
		quitting = true;
	});
	app.on('will-quit', () => backend?.stop());
	if (process.platform !== 'darwin') {
		// The default menu (File, Edit, ...) has nothing to offer here; macOS keeps it for copy and paste.
		Menu.setApplicationMenu(null);
	}
	void app.whenReady().then(start);
}
