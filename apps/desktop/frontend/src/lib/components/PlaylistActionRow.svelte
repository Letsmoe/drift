<script lang="ts">
	import type { Snippet } from 'svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { isContextPlaying } from '#lib/player/contextPlayback.ts';

	let {
		context,
		title,
		onplay,
		children = undefined
	}: {
		context: string;
		title: string;
		onplay: () => void;
		/** Save, follow and "..." controls next to the play button. */
		children?: Snippet;
	} = $props();
</script>

<div class="flex items-center gap-4 py-4">
	<PlayOverlayButton
		size="lg"
		playing={isContextPlaying(context)}
		label="Play {title}"
		onclick={onplay}
	/>
	{@render children?.()}
</div>
