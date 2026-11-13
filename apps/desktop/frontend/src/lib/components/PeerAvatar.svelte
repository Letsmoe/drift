<!-- A round tile with the first letter of a person's name. -->
<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import Label from '#lib/components/Label.svelte';
	import { colors } from '#lib/theme.ts';

	let { name, highlighted = false }: { name: string; highlighted?: boolean } = $props();

	const initial = $derived(name.trim().charAt(0).toUpperCase());
	const background = $derived(pickBackground());
	const textColor = $derived(pickTextColor());

	function pickBackground(): string {
		if (highlighted) {
			return colors.action;
		}
		return colors.raised;
	}

	function pickTextColor(): string {
		if (highlighted) {
			return colors.onAction;
		}
		return colors.text;
	}
</script>

<View
	style={{
		size: 32,
		shrink: 0,
		radius: 9999,
		align: 'center',
		justify: 'center',
		background,
		color: textColor
	}}
>
	<Label size="body" weight="semibold">{initial}</Label>
</View>
