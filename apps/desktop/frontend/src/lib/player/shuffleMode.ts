import ShuffleIcon from 'phosphor-svelte/lib/ShuffleIcon';
import type { QueueState } from '#lib/api/queue.ts';
import MagicShuffleIcon from '#lib/components/MagicShuffleIcon.svelte';

/** Tooltip and accessible name of the shuffle button, which cycles off, shuffle, magic shuffle. */
export function shuffleLabel(queue: QueueState): string {
	if (queue.magicShuffle) {
		return 'Magic shuffle: on, mixes in recommended songs';
	}
	if (queue.shuffle) {
		return 'Shuffle: on. Click again for magic shuffle';
	}
	return 'Shuffle: off';
}

export function shuffleIcon(queue: QueueState) {
	if (queue.magicShuffle) {
		return MagicShuffleIcon;
	}
	return ShuffleIcon;
}
