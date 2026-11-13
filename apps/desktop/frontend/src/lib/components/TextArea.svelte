<!--
  A multi-line text field in Drift's look. Like TextInput it tells textFocus while it has the keyboard, so the
  Space shortcut leaves the typing alone.
-->
<script lang="ts">
	import { Input, type Style } from '@neoworks/svelte-gpui/ui';
	import { textFocus } from '#lib/layout/textFocus.svelte.ts';
	import { colors, fontSizes, radii } from '#lib/theme.ts';

	let {
		value = $bindable(''),
		placeholder,
		rows = 4,
		mono = false,
		style
	}: {
		value?: string;
		placeholder?: string;
		rows?: number;
		mono?: boolean;
		style?: Style;
	} = $props();

	let focused = false;

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
	}

	function fontFamily(): 'mono' | 'sans' {
		if (mono) {
			return 'mono';
		}
		return 'sans';
	}
</script>

<Input
	bind:value
	multiline
	{rows}
	{placeholder}
	style={{
		minWidth: 0,
		paddingX: 12,
		paddingY: 8,
		radius: radii.medium,
		background: colors.canvas,
		color: colors.text,
		fontSize: fontSizes.tiny,
		lineHeight: 18,
		fontFamily: fontFamily(),
		...style
	}}
	onfocus={handleFocus}
	onblur={handleBlur}
/>
