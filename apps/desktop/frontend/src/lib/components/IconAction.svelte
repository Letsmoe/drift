<!--
  A bare icon that can be pressed. The old app's icon buttons were this: a colour that brightens on hover,
  sometimes with a round wash behind it.
-->
<script lang="ts">
	import { Button, Icon, type PointerEvent, type Style } from '@neoworks/svelte-gpui/ui';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	let {
		icon,
		iconSize = 20,
		boxSize,
		color = colors.muted,
		hoverColor = colors.text,
		hoverBackground,
		disabled = false,
		style,
		onpress
	}: {
		/** SVG markup. */
		icon: string;
		iconSize?: number;
		/** Width and height of the pressable area; the icon's size when left out. */
		boxSize?: number;
		color?: string;
		hoverColor?: string;
		/** A wash behind the icon on hover, as `hover:bg-hover` was. */
		hoverBackground?: string;
		disabled?: boolean;
		style?: Style;
		onpress?: (event: PointerEvent) => void;
	} = $props();

	const side = $derived(sideOf());
	const hoverLook = $derived(buildHoverLook());

	function sideOf(): number {
		if (boxSize === undefined) {
			return iconSize;
		}
		return boxSize;
	}

	function buildHoverLook(): Style {
		const built: Style = { color: hoverColor };
		if (hoverBackground !== undefined) {
			built.background = hoverBackground;
		}
		return built;
	}
</script>

<Button
	style={{
		size: side,
		radius: radii.full,
		align: 'center',
		justify: 'center',
		shrink: 0,
		color,
		transition: hoverTransition,
		...style
	}}
	hover={hoverLook}
	{disabled}
	{onpress}
>
	<Icon svg={icon} size={iconSize} />
</Button>
