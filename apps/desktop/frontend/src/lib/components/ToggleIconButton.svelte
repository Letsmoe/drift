<script lang="ts">
	import { Button, Icon, Tooltip, View } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	let {
		icon,
		activeIcon,
		label,
		active,
		onpress
	}: {
		/** SVG markup. */
		icon: string;
		/** Drawn instead of `icon` while active, such as a bolder weight. */
		activeIcon?: string;
		label: string;
		active: boolean;
		onpress: () => void;
	} = $props();

	const shownIcon = $derived(pickIcon());
	const color = $derived(pickColor());

	function pickIcon(): string {
		if (active && activeIcon !== undefined) {
			return activeIcon;
		}
		return icon;
	}

	function pickColor(): string {
		if (active) {
			return appearance.accent;
		}
		return appearance.muted;
	}

	function hoverColor(): string {
		if (active) {
			return appearance.accent;
		}
		return colors.text;
	}
</script>

<Tooltip text={label}>
	<Button
		style={{
			size: 32,
			radius: radii.full,
			align: 'center',
			justify: 'center',
			shrink: 0,
			color,
			transition: hoverTransition
		}}
		hover={{ background: colors.hover, color: hoverColor() }}
		{onpress}
	>
		<Icon svg={shownIcon} size={18} />
		{#if active}
			<View
				style={{
					position: 'absolute',
					bottom: 2,
					size: 4,
					radius: radii.full,
					background: appearance.accent
				}}
			/>
		{/if}
	</Button>
</Tooltip>
