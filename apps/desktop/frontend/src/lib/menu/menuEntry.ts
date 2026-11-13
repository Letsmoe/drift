export interface MenuEntry {
	label: string;
	/** SVG markup of the icon before the label. */
	icon?: string;
	onselect?: () => void;
	/** Entries that open a submenu instead of acting. */
	children?: MenuEntry[];
	separatorBefore?: boolean;
}
