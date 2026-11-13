import type { PointerEvent } from '@neoworks/svelte-gpui/ui';

export interface MenuPosition {
	x: number;
	y: number;
}

/** Open/closed state and anchor point of one popup menu. */
export class PopupMenu {
	position = $state<MenuPosition | null>(null);

	get isOpen(): boolean {
		return this.position !== null;
	}

	/** Opens the menu at a point in the window. */
	openAt(x: number, y: number): void {
		this.position = { x, y };
	}

	/** Opens the menu where a right press happened. */
	openAtPointer(event: PointerEvent): void {
		this.openAt(event.windowX, event.windowY);
	}

	close(): void {
		this.position = null;
	}
}
