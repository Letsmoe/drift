<script lang="ts">
	import { onDestroy } from 'svelte';
	import { LoadingSpinner } from '@neoworks-dev/ui';
	import Dialog from '#lib/components/Dialog.svelte';
	import ImportPreviewStep from '#lib/components/ImportPreviewStep.svelte';
	import ImportProgressStep from '#lib/components/ImportProgressStep.svelte';
	import ImportSourceStep from '#lib/components/ImportSourceStep.svelte';
	import ImportSummaryStep from '#lib/components/ImportSummaryStep.svelte';
	import { SpotifyImportFlow } from '#lib/import/spotifyImport.svelte.ts';

	let { onclose }: { onclose: () => void } = $props();

	const flow = new SpotifyImportFlow();

	onDestroy(() => flow.dispose());
</script>

<Dialog title="Import from Spotify" wide {onclose}>
	{#if flow.step === 'source'}
		<ImportSourceStep {flow} />
	{:else if flow.step === 'loading'}
		<div class="flex justify-center py-10"><LoadingSpinner /></div>
	{:else if flow.step === 'preview' && flow.summary}
		<ImportPreviewStep {flow} summary={flow.summary} />
	{:else if flow.step === 'running'}
		<ImportProgressStep {flow} />
	{:else if flow.step === 'done' && flow.result}
		<ImportSummaryStep {flow} result={flow.result} {onclose} />
	{/if}
</Dialog>
