<script lang="ts">
	import type { Snippet } from 'svelte';
	import ShuffleIcon from 'phosphor-svelte/lib/ShuffleIcon';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { isContextPlaying } from '#lib/player/contextPlayback.ts';
	import { rememberContextTitle } from '#lib/queue/contextLabel.ts';

	let {
		context,
		title,
		onplay,
		onshuffle = undefined,
		children = undefined
	}: {
		context: string;
		title: string;
		onplay: () => void;
		/** Shuffle play; the button only shows when set. */
		onshuffle?: () => void;
		/** Save, follow and "..." controls next to the play button. */
		children?: Snippet;
	} = $props();

	let rowVisible = $state(true);

	$effect(() => {
		rememberContextTitle(context, title);
	});

	// The page scrolls inside its pane, which the implicit root accounts for by clipping.
	function watchVisibility(row: HTMLElement) {
		const observer = new IntersectionObserver(([entry]) => {
			rowVisible = entry.isIntersecting;
		});
		observer.observe(row);
		return () => observer.disconnect();
	}
</script>

<!-- Zero-height sticky anchor: the bar overlays the page top only while the row is out of view. -->
<div class="sticky top-0 z-30 -mx-6 h-0">
	{#if !rowVisible}
		<div
			class="flex h-16 items-center gap-3 rounded-t-2xl border-b border-line-faint bg-elevated/90 px-6 shadow-lg backdrop-blur"
		>
			<PlayOverlayButton
				size="sm"
				playing={isContextPlaying(context)}
				label="Play {title}"
				onclick={onplay}
			/>
			<p class="truncate text-lg font-bold">{title}</p>
		</div>
	{/if}
</div>

<div class="flex items-center gap-4 py-4" {@attach watchVisibility}>
	<PlayOverlayButton
		size="lg"
		playing={isContextPlaying(context)}
		label="Play {title}"
		onclick={onplay}
	/>
	{#if onshuffle}
		<button
			type="button"
			aria-label="Shuffle play {title}"
			title="Shuffle play"
			onclick={onshuffle}
			class="duration-fast flex size-10 items-center justify-center rounded-full text-muted transition-colors hover:bg-hover hover:text-default"
		>
			<ShuffleIcon size={24} />
		</button>
	{/if}
	{@render children?.()}
</div>
