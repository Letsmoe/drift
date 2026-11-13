<!-- The kit's toggle icon button, with Drift's colours: the cover's accent while on, the secondary text colour off. -->
<script lang="ts">
	import type { Style } from '@neoworks/svelte-gpui/ui';
	import { IconButton } from '@neoworks/svelte-gpui-ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import { colors } from '#lib/theme.ts';

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

	const look = $derived(pickLook());

	// A raised background keeps the state readable when the cover's accent is close to the muted text.
	function pickLook(): Style {
		if (active) {
			return { color: appearance.accent, background: colors.raised, shrink: 0 };
		}
		return { color: appearance.muted, shrink: 0 };
	}
</script>

<IconButton
	size="sm"
	{icon}
	selectedIcon={activeIcon}
	{label}
	selected={active}
	style={look}
	{onpress}
/>
