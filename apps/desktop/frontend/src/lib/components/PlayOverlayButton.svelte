<script lang="ts">
	import { Button, Icon, type PointerEvent, type Style } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import { pauseFill, playFill } from '#lib/icons.ts';
	import { hoverTransition, pressedSolidLook, radii } from '#lib/theme.ts';

	type Size = 'sm' | 'md' | 'lg';

	let {
		playing,
		size = 'md',
		style,
		onpress
	}: {
		playing: boolean;
		size?: Size;
		/** Position and opacity, such as the corner of a card. */
		style?: Style;
		onpress: (event: PointerEvent) => void;
	} = $props();

	const sides: Record<Size, number> = { sm: 32, md: 48, lg: 49 };
	const iconSizes: Record<Size, number> = { sm: 14, md: 20, lg: 24 };

	const icon = $derived(iconFor(playing));

	function iconFor(isPlaying: boolean): string {
		if (isPlaying) {
			return pauseFill;
		}
		return playFill;
	}
</script>

<Button
	style={{
		size: sides[size],
		radius: radii.full,
		align: 'center',
		justify: 'center',
		shrink: 0,
		background: appearance.accent,
		color: appearance.onAccent,
		shadow: { y: 4, blur: 12, color: 'rgba(0, 0, 0, 0.45)' },
		transition: hoverTransition,
		...style
	}}
	hover={{ background: appearance.accentHover }}
	active={pressedSolidLook}
	{onpress}
>
	<Icon svg={icon} size={iconSizes[size]} />
</Button>
