<script lang="ts">
	import { Button } from '@neoworks-dev/ui';
	import type { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';

	let { flow }: { flow: MusicImportFlow } = $props();
</script>

<div class="flex flex-col gap-4">
	<div class="flex items-baseline justify-between text-sm">
		<span class="font-semibold">{flow.phaseLabel}</span>
		{#if flow.progress}
			<span class="text-muted">{flow.progress.done} / {flow.progress.total}</span>
		{/if}
	</div>
	<div
		class="h-2 overflow-hidden rounded-full bg-raised"
		role="progressbar"
		aria-valuemin="0"
		aria-valuemax="100"
		aria-valuenow={Math.round(flow.progressFraction * 100)}
	>
		<div
			class="h-full rounded-full bg-action transition-all"
			style:width="{flow.progressFraction * 100}%"
		></div>
	</div>
	<p class="min-h-5 truncate text-sm text-muted">
		{#if flow.progress}{flow.progress.current}{/if}
	</p>
	<div class="flex justify-end gap-2">
		<Button variant="ghost" onclick={() => flow.cancel()}>Cancel</Button>
		<Button onclick={() => flow.closeDialog()}>Continue in background</Button>
	</div>
</div>
