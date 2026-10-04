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

	openAt(x: number, y: number): void {
		this.position = { x, y };
	}

	/** Opens the menu under an element, e.g. its "..." button. */
	openBelow(element: HTMLElement): void {
		const bounds = element.getBoundingClientRect();
		this.openAt(bounds.left, bounds.bottom + 4);
	}

	close(): void {
		this.position = null;
	}
}
