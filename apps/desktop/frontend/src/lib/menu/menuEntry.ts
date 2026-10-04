import type { Component } from 'svelte';

export interface MenuEntry {
	label: string;
	icon?: Component<{ size?: number | string }>;
	onselect?: () => void;
	/** Entries that open a submenu instead of acting. */
	children?: MenuEntry[];
	separatorBefore?: boolean;
}
