import { getItem, setItem } from '#lib/storage.ts';

const storageKey = 'drift:import-prompt-hidden';

function loadHidden(): boolean {
	return getItem(storageKey) === 'true';
}

/** Whether the library sidebar still offers the "Import your music" card. */
class ImportPrompt {
	hidden = $state(loadHidden());

	hide(): void {
		this.hidden = true;
		setItem(storageKey, 'true');
	}
}

export const importPrompt = new ImportPrompt();
