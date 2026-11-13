<!--
  The frame of an album, playlist or podcast page: a tinted wash under the header that fades out, with the
  page's own scrolling content over it. The wash stays at the top of the pane while the content scrolls.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import { View } from '@neoworks/svelte-gpui/ui';
	import VerticalFade from '#lib/components/VerticalFade.svelte';
	import { gradientColorFor } from '#lib/pageColor.ts';

	const washHeight = 416;

	let { seed, children }: { seed: string; children: Snippet } = $props();

	const color = $derived(gradientColorFor(seed));
</script>

<View style={{ flex: 1, minHeight: 0, position: 'relative' }}>
	<VerticalFade {color} height={washHeight} />
	{@render children()}
</View>
