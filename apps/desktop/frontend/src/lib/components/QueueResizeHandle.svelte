<script lang="ts">
	import { View, type KeyEvent, type PointerEvent } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import { panels, queueWidthLimits } from '#lib/layout/panels.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	const keyboardStep = 16;
	const handleWidth = 8;
	const doubleClick = 2;

	// Pointer position (in the window) and panel width when the drag started; the panel grows as the pointer moves left.
	let dragStart = $state<{ pointerX: number; width: number } | undefined>(undefined);
	let hovered = $state(false);

	const barColor = $derived(pickBarColor());

	function pickBarColor(): string {
		if (dragStart) {
			return appearance.accent;
		}
		if (hovered) {
			return colors.lineStrong;
		}
		return 'transparent';
	}

	function startDrag(event: PointerEvent): void {
		if (event.clicks === doubleClick) {
			panels.resizeQueue(queueWidthLimits.initial);
			return;
		}
		dragStart = { pointerX: event.windowX, width: panels.queueWidth };
	}

	function drag(event: PointerEvent): void {
		if (!dragStart) {
			return;
		}
		panels.resizeQueue(dragStart.width + dragStart.pointerX - event.windowX);
	}

	function endDrag(): void {
		dragStart = undefined;
	}

	function resizeWithKeyboard(event: KeyEvent): void {
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

<View
	style={{
		position: 'absolute',
		top: 16,
		bottom: 16,
		left: -handleWidth,
		width: handleWidth,
		align: 'center',
		cursor: 'col-resize'
	}}
	focusable
	onhover={(isHovered) => (hovered = isHovered)}
	ondown={startDrag}
	onmove={drag}
	onup={endDrag}
	onkey={resizeWithKeyboard}
>
	<View
		style={{
			width: 2,
			flex: 1,
			radius: radii.full,
			background: barColor,
			transition: hoverTransition
		}}
	/>
</View>
