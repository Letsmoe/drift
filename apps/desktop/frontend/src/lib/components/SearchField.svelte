<script lang="ts">
	import { Button, Icon, Input, View, type KeyEvent } from '@neoworks/svelte-gpui/ui';
	import type { Snippet } from 'svelte';
	import { appearance } from '#lib/appearance.svelte.ts';
	import { magnifyingGlass, x } from '#lib/layout/chromeIcons.ts';
	import { textFocus } from '#lib/layout/textFocus.svelte.ts';
	import { colors, fontSizes, glassBackground, hoverTransition, radii } from '#lib/theme.ts';

	let {
		value = $bindable(''),
		placeholder = 'Search',
		onsubmit,
		onfocus,
		oninput,
		onblur,
		onkey,
		children
	}: {
		value?: string;
		placeholder?: string;
		onsubmit: () => void;
		onfocus?: () => void;
		oninput?: () => void;
		onblur?: () => void;
		onkey?: (event: KeyEvent) => void;
		/** Drawn against the field's box, such as a suggestions list. */
		children?: Snippet;
	} = $props();

	let field = $state<Input>();
	let focused = $state(false);
	let hovered = $state(false);

	const borderColor = $derived(pickBorderColor());
	const background = $derived(pickBackground());

	function pickBorderColor(): string {
		if (focused) {
			return colors.lineStrong;
		}
		if (hovered) {
			return colors.line;
		}
		return 'transparent';
	}

	function pickBackground(): string {
		if (hovered && !focused) {
			return colors.raised;
		}
		return glassBackground;
	}

	function handleFocus(): void {
		focused = true;
		textFocus.enter();
		onfocus?.();
	}

	function handleBlur(): void {
		focused = false;
		textFocus.leave();
		onblur?.();
	}

	function clear(): void {
		value = '';
		oninput?.();
		field?.focus();
	}
</script>

<View
	style={{
		direction: 'row',
		align: 'center',
		height: 48,
		gap: 12,
		paddingX: 16,
		radius: radii.full,
		background,
		borderWidth: 1,
		borderColor,
		transition: hoverTransition
	}}
	onhover={(isHovered) => (hovered = isHovered)}
>
	<Icon svg={magnifyingGlass} size={20} color={appearance.dim} style={{ shrink: 0 }} />
	<Input
		bind:this={field}
		bind:value
		{placeholder}
		style={{
			flex: 1,
			minWidth: 0,
			alignSelf: 'center',
			fontSize: fontSizes.medium,
			color: colors.text
		}}
		onfocus={handleFocus}
		onblur={handleBlur}
		oninput={() => oninput?.()}
		onsubmit={() => onsubmit()}
		{onkey}
	/>
	{#if value !== ''}
		<Button
			style={{ shrink: 0, color: appearance.dim }}
			hover={{ color: colors.text }}
			onpress={clear}
		>
			<Icon svg={x} size={18} />
		</Button>
	{/if}
	{@render children?.()}
</View>
