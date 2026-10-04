/** A stable, muted colour for a page's header gradient, derived from its title. */
export function gradientColorFor(seed: string): string {
	let hash = 0;
	for (const character of seed) {
		hash = (hash * 31 + character.charCodeAt(0)) % 360;
	}
	return `hsl(${hash} 40% 30%)`;
}
