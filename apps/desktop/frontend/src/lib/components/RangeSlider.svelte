<script lang="ts">
	import { View, type PointerEvent } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import { colors, radii } from '#lib/theme.ts';

	// Music-specific for now (seek, volume).
	let {
		value,
		max,
		step = 0,
		disabled = false,
		oninput,
		oncommit
	}: {
		value: number;
		max: number;
		/** Values snap to multiples of this; 0 leaves them free. */
		step?: number;
		disabled?: boolean;
		/** Called with every value while the thumb is dragged, and once on a plain press. */
		oninput?: (value: number) => void;
		/** Called with the value the thumb was released at. */
		oncommit?: (value: number) => void;
	} = $props();

	const hitHeight = 16;
	const trackHeight = 4;
	const thumbSize = 12;

	let width = $state(0);
	let hovered = $state(false);
	let dragValue = $state<number | undefined>(undefined);

	const shownValue = $derived(pickShownValue());
	const fillWidth = $derived(fillWidthOf(shownValue));
	const thumbVisible = $derived((hovered || dragValue !== undefined) && !disabled);

	function pickShownValue(): number {
		if (dragValue !== undefined) {
			return dragValue;
		}
		return value;
	}

	function fillWidthOf(shown: number): number {
		if (max <= 0) {
			return 0;
		}
		return Math.min(width, Math.max(0, (shown / max) * width));
	}

	function valueAt(pointerX: number): number {
		if (width <= 0 || max <= 0) {
			return 0;
		}
		let next = Math.min(max, Math.max(0, (pointerX / width) * max));
		if (step > 0) {
			next = Math.round(next / step) * step;
		}
		return next;
	}

	function startDrag(event: PointerEvent): void {
		if (disabled) {
			return;
		}
		dragValue = valueAt(event.x);
		oninput?.(dragValue);
	}

	function moveDrag(event: PointerEvent): void {
		if (dragValue === undefined) {
			return;
		}
		dragValue = valueAt(event.x);
		oninput?.(dragValue);
	}

	function endDrag(): void {
		if (dragValue === undefined) {
			return;
		}
		const released = dragValue;
		dragValue = undefined;
		oncommit?.(released);
	}
</script>

<View
	style={{ flex: 1, height: hitHeight, justify: 'center' }}
	onlayout={(layout) => (width = layout.width)}
	onhover={(isHovered) => (hovered = isHovered)}
	ondown={startDrag}
	onmove={moveDrag}
	onup={endDrag}
>
	<View
		style={{
			height: trackHeight,
			radius: radii.full,
			background: colors.hover
		}}
	>
		<View
			style={{
				width: fillWidth,
				height: trackHeight,
				radius: radii.full,
				background: appearance.accent
			}}
		/>
	</View>
	{#if thumbVisible}
		<View
			style={{
				position: 'absolute',
				left: Math.max(0, fillWidth - thumbSize / 2),
				top: (hitHeight - thumbSize) / 2,
				size: thumbSize,
				radius: radii.full,
				background: colors.text
			}}
		/>
	{/if}
</View>
