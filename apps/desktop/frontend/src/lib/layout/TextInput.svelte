<!--
  A single-line text field in Drift's look (a raised well). It tells textFocus while it has the keyboard,
  so the Space shortcut leaves the typing alone, and lets go when it is removed while focused.
-->
<script lang="ts">
	import { Input, type KeyEvent, type Style } from '@neoworks/svelte-gpui/ui';
	import { textFocus } from '#lib/layout/textFocus.svelte.ts';
	import { colors, fontSizes, radii } from '#lib/theme.ts';

	let {
		value = $bindable(''),
		placeholder,
		maxLength,
		autofocus = false,
		mono = false,
		height = 36,
		style,
		onsubmit,
		onescape,
		onblur
	}: {
		value?: string;
		placeholder?: string;
		maxLength?: number;
		/** Takes the keyboard as soon as it appears. */
		autofocus?: boolean;
		mono?: boolean;
		height?: number;
		/** Layout only, such as a width. */
		style?: Style;
		onsubmit?: (value: string) => void;
		onescape?: () => void;
		onblur?: () => void;
	} = $props();

	const textLineHeight = 20;

	let field: Input | undefined = $state();
	let focused = false;

	$effect(() => {
		if (autofocus && field !== undefined) {
			field.focus();
		}
	});

	$effect(() => {
		return () => {
			if (focused) {
				textFocus.leave();
				focused = false;
			}
		};
	});

	function handleFocus(): void {
		if (!focused) {
			focused = true;
			textFocus.enter();
		}
	}

	function handleBlur(): void {
		if (focused) {
			focused = false;
			textFocus.leave();
		}
		onblur?.();
	}

	function handleKey(event: KeyEvent): void {
		if (event.key === 'Escape' && onescape !== undefined) {
			event.stopPropagation();
			onescape();
		}
	}

	function fontFamily(): 'mono' | 'sans' {
		if (mono) {
			return 'mono';
		}
		return 'sans';
	}
</script>

<Input
	bind:this={field}
	bind:value
	{placeholder}
	{maxLength}
	style={{
		minWidth: 0,
		paddingX: 12,
		paddingY: (height - textLineHeight) / 2,
		lineHeight: textLineHeight,
		radius: radii.medium,
		background: colors.raised,
		color: colors.text,
		fontSize: fontSizes.small,
		fontFamily: fontFamily(),
		...style
	}}
	focus={{ background: colors.hover }}
	onfocus={handleFocus}
	onblur={handleBlur}
	onkey={handleKey}
	{onsubmit}
/>
