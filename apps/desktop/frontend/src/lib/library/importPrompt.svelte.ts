const storageKey = 'otter-music:import-prompt-hidden';

function loadHidden(): boolean {
	try {
		return localStorage.getItem(storageKey) === 'true';
	} catch {
		return false;
	}
}

/** Whether the library sidebar still offers the "Import your music" card. */
class ImportPrompt {
	hidden = $state(loadHidden());

	hide(): void {
		this.hidden = true;
		localStorage.setItem(storageKey, 'true');
	}
}

export const importPrompt = new ImportPrompt();
