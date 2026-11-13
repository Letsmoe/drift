<!--
  The bar that overlays the top of a page once its play button has scrolled out of view. The page tells it
  how far it has scrolled; the old one watched the row with an intersection observer.
-->
<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import Label from '#lib/components/Label.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { isContextPlaying } from '#lib/player/contextPlayback.ts';
	import { colors, radii } from '#lib/theme.ts';

	let {
		visible,
		context,
		title,
		onplay
	}: { visible: boolean; context: string; title: string; onplay: () => void } = $props();
</script>

{#if visible}
	<View
		style={{
			position: 'absolute',
			top: 0,
			left: 0,
			right: 0,
			height: 64,
			direction: 'row',
			align: 'center',
			gap: 12,
			paddingX: 24,
			background: 'rgba(28, 28, 30, 0.92)',
			borderBottomWidth: 1,
			borderColor: colors.lineFaint,
			radius: [radii.huge, radii.huge, 0, 0],
			shadow: { y: 12, blur: 32, color: 'rgba(0, 0, 0, 0.55)' }
		}}
	>
		<PlayOverlayButton size="sm" playing={isContextPlaying(context)} onpress={onplay} />
		<Label size="large" weight="bold" truncate style={{ flex: 1, minWidth: 0 }}>{title}</Label>
	</View>
{/if}
