import type { QueueState } from '#lib/api/queue.ts';
import { shuffle, shuffleBold, sparkleFill } from '#lib/layout/chromeIcons.ts';

// The inside of a Phosphor SVG: what sits between the <svg> tags.
function innerMarkup(svg: string): string {
	const start = svg.indexOf('>') + 1;
	const end = svg.lastIndexOf('</svg>');
	return svg.slice(start, end);
}

/** Magic shuffle: the shuffle arrows with a sparkle in the corner. */
function composeMagicShuffle(arrows: string): string {
	return (
		'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" fill="currentColor">' +
		`<g transform="translate(0 36) scale(0.85)">${innerMarkup(arrows)}</g>` +
		`<g transform="translate(138 -6) scale(0.46)">${innerMarkup(sparkleFill)}</g>` +
		'</svg>'
	);
}

const magicShuffle = composeMagicShuffle(shuffle);
const magicShuffleBold = composeMagicShuffle(shuffleBold);

/** Tooltip of the shuffle button, which cycles off, shuffle, magic shuffle. */
export function shuffleLabel(queue: QueueState): string {
	if (queue.magicShuffle) {
		return 'Magic shuffle: on, mixes in recommended songs';
	}
	if (queue.shuffle) {
		return 'Shuffle: on. Click again for magic shuffle';
	}
	return 'Shuffle: off';
}

export function shuffleIcon(queue: QueueState): string {
	if (queue.magicShuffle) {
		return magicShuffle;
	}
	return shuffle;
}

export function shuffleActiveIcon(queue: QueueState): string {
	if (queue.magicShuffle) {
		return magicShuffleBold;
	}
	return shuffleBold;
}
