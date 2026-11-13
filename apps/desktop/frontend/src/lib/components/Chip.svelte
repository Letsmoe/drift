<script lang="ts">
	import { Button } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Label from '#lib/components/Label.svelte';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	let {
		label,
		selected = false,
		paddingX = 12,
		onclick
	}: { label: string; selected?: boolean; paddingX?: number; onclick: () => void } = $props();

	const background = $derived(backgroundFor(selected));
	const textColor = $derived(textColorFor(selected));
	const hoverBackground = $derived(hoverBackgroundFor(selected));

	function backgroundFor(isSelected: boolean): string {
		if (isSelected) {
			return appearance.accent;
		}
		return colors.raised;
	}

	function textColorFor(isSelected: boolean): string {
		if (isSelected) {
			return appearance.onAccent;
		}
		return colors.text;
	}

	function hoverBackgroundFor(isSelected: boolean): string {
		if (isSelected) {
			return appearance.accentHover;
		}
		return colors.hover;
	}
</script>

<Button
	style={{
		height: 32,
		paddingX,
		radius: radii.full,
		shrink: 0,
		align: 'center',
		justify: 'center',
		background,
		color: textColor,
		transition: hoverTransition
	}}
	hover={{ background: hoverBackground }}
	onpress={onclick}
>
	<Label size="small" weight="medium" style={{ nowrap: true }}>{label}</Label>
</Button>
