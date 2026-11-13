<script lang="ts">
	import { Icon } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { arrowCircleDownFill } from '#lib/icons.ts';

	let { videoId, size = 14 }: { videoId: string; size?: number } = $props();
</script>

{#if downloads.isDownloaded(videoId)}
	<Icon svg={arrowCircleDownFill} {size} color={appearance.accent} style={{ shrink: 0 }} />
{/if}
