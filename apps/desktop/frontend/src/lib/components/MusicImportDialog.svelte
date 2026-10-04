<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { LoadingSpinner } from '@neoworks-dev/ui';
	import Dialog from '#lib/components/Dialog.svelte';
	import ImportPreviewStep from '#lib/components/ImportPreviewStep.svelte';
	import ImportProgressStep from '#lib/components/ImportProgressStep.svelte';
	import ImportSourceStep from '#lib/components/ImportSourceStep.svelte';
	import ImportSummaryStep from '#lib/components/ImportSummaryStep.svelte';
	import YouTubeSignInStep from '#lib/components/YouTubeSignInStep.svelte';
	import { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';

	let { onclose }: { onclose: () => void } = $props();

	const flow = new MusicImportFlow();

	onMount(() => flow.youtube.refresh());
	onDestroy(() => flow.dispose());
</script>

<Dialog title="Import music" wide {onclose}>
	{#if flow.step === 'source'}
		<ImportSourceStep {flow} />
	{:else if flow.step === 'signin'}
		<YouTubeSignInStep {flow} />
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
