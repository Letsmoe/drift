import type { MenuEntry as KitMenuEntry } from '@neoworks/svelte-gpui-ui';
import type { MenuEntry } from '#lib/menu/menuEntry.ts';

/** Drift's flat menu entries as the kit's `Menu` takes them. Entries with submenus are not supported by the kit. */
export function toKitMenuItems(entries: MenuEntry[]): KitMenuEntry[] {
	const items: KitMenuEntry[] = [];
	for (const entry of entries) {
		if (entry.separatorBefore) {
			items.push({ type: 'divider' });
		}
		items.push({ label: entry.label, icon: entry.icon, onselect: entry.onselect });
	}
	return items;
}
