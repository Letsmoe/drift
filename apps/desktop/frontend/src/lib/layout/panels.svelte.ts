/** Which side panels are open; shared by the player bar toggle and the layout. */
class Panels {
	queueOpen = $state(true);
	/** The full-screen now playing view covers the whole window. */
	fullPlayerOpen = $state(false);

	toggleQueue(): void {
		this.queueOpen = !this.queueOpen;
	}

	toggleFullPlayer(): void {
		this.fullPlayerOpen = !this.fullPlayerOpen;
	}

	closeFullPlayer(): void {
		this.fullPlayerOpen = false;
	}
}

export const panels = new Panels();
