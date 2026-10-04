<script lang="ts">
	import ArrowCircleDownIcon from 'phosphor-svelte/lib/ArrowCircleDownIcon';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';

	let { videoId, size = 14 }: { videoId: string; size?: number } = $props();
</script>

{#if downloads.isDownloaded(videoId)}
	<span title="Downloaded" class="shrink-0 text-green">
		<ArrowCircleDownIcon {size} weight="fill" />
		<span class="sr-only">Downloaded</span>
	</span>
{/if}
