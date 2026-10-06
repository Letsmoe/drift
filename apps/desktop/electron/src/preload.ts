import { contextBridge, ipcRenderer } from 'electron';

interface OpenDialogOptions {
	title: string;
	filters?: { name: string; extensions: string[] }[];
}

contextBridge.exposeInMainWorld('otter', {
	backend: () => ipcRenderer.sendSync('otter:backend'),
	pickFile: (options: OpenDialogOptions) => ipcRenderer.invoke('otter:pick-file', options),
	pickFolder: (options: OpenDialogOptions) => ipcRenderer.invoke('otter:pick-folder', options)
});
