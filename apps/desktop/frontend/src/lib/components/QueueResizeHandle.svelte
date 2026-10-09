<script lang="ts">
	import { panels, queueWidthLimits } from '#lib/layout/panels.svelte.ts';

	const keyboardStep = 16;

	// Pointer position and panel width when the drag started; the panel grows as the pointer moves left.
	let dragStart: { pointerX: number; width: number } | undefined = $state(undefined);

	function startDrag(event: PointerEvent): void {
		event.preventDefault();
		const handle = event.currentTarget as HTMLElement;
		handle.setPointerCapture(event.pointerId);
		dragStart = { pointerX: event.clientX, width: panels.queueWidth };
	}

	function drag(event: PointerEvent): void {
		if (!dragStart) {
			return;
		}
		panels.resizeQueue(dragStart.width + dragStart.pointerX - event.clientX);
	}

	function endDrag(): void {
		dragStart = undefined;
	}

	function resizeWithKeyboard(event: KeyboardEvent): void {
		if (event.key === 'ArrowLeft') {
			event.preventDefault();
			panels.resizeQueue(panels.queueWidth + keyboardStep);
		}
		if (event.key === 'ArrowRight') {
			event.preventDefault();
			panels.resizeQueue(panels.queueWidth - keyboardStep);
		}
	}
</script>

<!-- A focusable separator is the ARIA pattern for a resizable splitter. -->
<!-- svelte-ignore a11y_no_noninteractive_tabindex, a11y_no_noninteractive_element_interactions -->
<div
	role="separator"
	aria-orientation="vertical"
	aria-label="Resize Now playing"
	aria-valuenow={panels.queueWidth}
	aria-valuemin={queueWidthLimits.minimum}
	aria-valuemax={queueWidthLimits.maximum}
	tabindex="0"
	onpointerdown={startDrag}
	onpointermove={drag}
	onpointerup={endDrag}
	onpointercancel={endDrag}
	ondblclick={() => panels.resizeQueue(queueWidthLimits.initial)}
	onkeydown={resizeWithKeyboard}
	class="group absolute inset-y-4 -left-2 z-10 flex w-2 cursor-col-resize touch-none justify-center outline-none"
>
	<span
		class={[
			'duration-fast w-0.5 rounded-full transition-colors group-hover:bg-line-strong group-focus-visible:bg-action',
			dragStart && 'bg-action'
		]}
	></span>
</div>
