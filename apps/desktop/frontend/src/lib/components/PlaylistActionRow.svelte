<script lang="ts">
	import type { Snippet } from 'svelte';
	import { View } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import IconAction from '#lib/components/IconAction.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { shuffle } from '#lib/icons.ts';
	import { isContextPlaying } from '#lib/player/contextPlayback.ts';
	import { rememberContextTitle } from '#lib/queue/contextLabel.ts';
	import { colors } from '#lib/theme.ts';

	let {
		context,
		title,
		onplay,
		onshuffle = undefined,
		children = undefined
	}: {
		context: string;
		title: string;
		onplay: () => void;
		/** Shuffle play; the button only shows when set. */
		onshuffle?: () => void;
		/** Save, follow and "..." controls next to the play button. */
		children?: Snippet;
	} = $props();

	$effect(() => {
		rememberContextTitle(context, title);
	});
</script>

<View style={{ direction: 'row', align: 'center', gap: 16, paddingY: 16 }}>
	<PlayOverlayButton size="lg" playing={isContextPlaying(context)} onpress={onplay} />
	{#if onshuffle}
		<IconAction
			icon={shuffle}
			iconSize={24}
			boxSize={40}
			color={appearance.muted}
			hoverColor={colors.text}
			hoverBackground={colors.hover}
			onpress={onshuffle}
		/>
	{/if}
	{@render children?.()}
</View>
