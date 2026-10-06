<script lang="ts">
	import CaretLeftIcon from 'phosphor-svelte/lib/CaretLeftIcon';
	import CaretRightIcon from 'phosphor-svelte/lib/CaretRightIcon';
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();

	// Pixels of slack so rounding at either end doesn't leave an arrow that does nothing.
	const edgeTolerance = 4;
	// Each click scrolls this share of the visible width, keeping a card of context.
	const pageShare = 0.85;

	let scroller: HTMLDivElement | undefined = $state();
	let canScrollBack = $state(false);
	let canScrollForward = $state(false);

	function updateArrows(): void {
		if (!scroller) {
			return;
		}
		canScrollBack = scroller.scrollLeft > edgeTolerance;
		const remaining = scroller.scrollWidth - scroller.clientWidth - scroller.scrollLeft;
		canScrollForward = remaining > edgeTolerance;
	}

	function scrollByPage(direction: 1 | -1): void {
		if (!scroller) {
			return;
		}
		scroller.scrollBy({ left: direction * scroller.clientWidth * pageShare, behavior: 'smooth' });
	}

	function observeSize(element: HTMLDivElement) {
		const observer = new ResizeObserver(updateArrows);
		observer.observe(element);
		if (element.firstElementChild) {
			observer.observe(element.firstElementChild);
		}
		return () => observer.disconnect();
	}

	function tabIndexFor(visible: boolean): number {
		if (visible) {
			return 0;
		}
		return -1;
	}
</script>

{#snippet arrow(direction: 1 | -1, label: string, visible: boolean)}
	<button
		type="button"
		aria-label={label}
		tabindex={tabIndexFor(visible)}
		onclick={() => scrollByPage(direction)}
		class={[
			'duration-fast absolute top-1/3 z-10 flex size-9 -translate-y-1/2 cursor-pointer items-center justify-center rounded-full bg-elevated text-default shadow-lg transition-opacity hover:bg-hover',
			direction === -1 && 'left-1',
			direction === 1 && 'right-1',
			visible && 'opacity-0 group-hover/scroll:opacity-100 focus-visible:opacity-100',
			!visible && 'pointer-events-none opacity-0'
		]}
	>
		{#if direction === -1}
			<CaretLeftIcon size={18} weight="bold" />
		{:else}
			<CaretRightIcon size={18} weight="bold" />
		{/if}
	</button>
{/snippet}

<div class="group/scroll relative">
	{@render arrow(-1, 'Scroll back', canScrollBack)}
	<div
		bind:this={scroller}
		onscroll={updateArrows}
		{@attach observeSize}
		class="[scrollbar-width:none] overflow-x-auto [&::-webkit-scrollbar]:hidden"
	>
		<div class="flex w-max">
			{@render children()}
		</div>
	</div>
	{@render arrow(1, 'Scroll forward', canScrollForward)}
</div>
