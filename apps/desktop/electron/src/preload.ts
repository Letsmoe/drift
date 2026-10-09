import { contextBridge, ipcRenderer } from 'electron';

interface OpenDialogOptions {
	title: string;
	filters?: { name: string; extensions: string[] }[];
}

contextBridge.exposeInMainWorld('drift', {
	backend: () => ipcRenderer.sendSync('drift:backend'),
	pickFile: (options: OpenDialogOptions) => ipcRenderer.invoke('drift:pick-file', options),
	pickFolder: (options: OpenDialogOptions) => ipcRenderer.invoke('drift:pick-folder', options)
});
