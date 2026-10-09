const queueWidthStorageKey = 'drift:queue-panel-width';

/** Bounds of the Now playing column's width in CSS pixels. */
export const queueWidthLimits = { minimum: 300, maximum: 560, initial: 360 };

function clampQueueWidth(width: number): number {
	return Math.min(queueWidthLimits.maximum, Math.max(queueWidthLimits.minimum, Math.round(width)));
}

function loadQueueWidth(): number {
	const stored = Number(localStorage.getItem(queueWidthStorageKey));
	if (!Number.isFinite(stored) || stored === 0) {
		return queueWidthLimits.initial;
	}
	return clampQueueWidth(stored);
}

/** Which side panels are open and how wide; shared by the player bar toggle and the layout. */
class Panels {
	queueOpen = $state(true);
	queueWidth = $state(loadQueueWidth());
	/** The full-screen now playing view covers the whole window. */
	fullPlayerOpen = $state(false);

	toggleQueue(): void {
		this.queueOpen = !this.queueOpen;
	}

	/** Sets the Now playing column's width, clamped to queueWidthLimits, and remembers it. */
	resizeQueue(width: number): void {
		this.queueWidth = clampQueueWidth(width);
		localStorage.setItem(queueWidthStorageKey, String(this.queueWidth));
	}

	toggleFullPlayer(): void {
		this.fullPlayerOpen = !this.fullPlayerOpen;
	}

	closeFullPlayer(): void {
		this.fullPlayerOpen = false;
	}
}

export const panels = new Panels();
