/** Which side panels are open; shared by the player bar toggle and the layout. */
class Panels {
	queueOpen = $state(true);

	toggleQueue(): void {
		this.queueOpen = !this.queueOpen;
	}
}

export const panels = new Panels();
